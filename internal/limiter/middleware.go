package limiter

import "net/http"

var _ http.Handler = (*Limiter)(nil)

type Limiter struct {
	// reqs is a buffed channel and used to track the number of currently executing requests. Each time the Limiter
	// starts processing a request, an empty struct is added, and when the request is finished, an empty struct is
	// pulled out.
	reqs chan struct{}

	// handler is the next http.Handler to run.
	handler       http.Handler
	dropOverLimit bool
}

// New constructs a Limiter middleware with specified maximum number of concurrent requests, handler to wrap, and
// a flag whether it should drop requests that exceed the limit number.
func New(maxReqs int, handler http.Handler, dropOverLimit bool) *Limiter {
	return &Limiter{
		reqs:          make(chan struct{}, maxReqs),
		handler:       handler,
		dropOverLimit: dropOverLimit,
	}
}

// ServeHTTP checks each request whether it exceeds the maximum number of concurrent requests. It has two options of
// handling this: drop the requests and return 503 status code, or wait for a different requests to finish and continue
// the execution.
func (l *Limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.dropOverLimit {
		// if limiter is configured to drop limit-exceeding requests, try to send an empty struct to the channel, and
		// if it is full, return 503
		select {
		case l.reqs <- struct{}{}:
		default:
			http.Error(w, "server is overloaded", http.StatusServiceUnavailable)
			return
		}
	} else {
		// otherwise, block the execution until there is a space in the channel
		l.reqs <- struct{}{}
	}
	l.handler.ServeHTTP(w, r)
	<-l.reqs // decrementing the number of empty structs in the channel to signal the end of request
}
