package limiter

import "net/http"

var _ http.Handler = (*Limiter)(nil)

type Limiter struct {
	// reqs is a buffed channel that tracks the number of currently executing requests. Each time the Limiter starts
	// processing a request, it adds an empty struct, and when the request is finished, it pulls out an empty struct.
	reqs chan struct{}

	// handler is the next http.Handler to run.
	handler       http.Handler
	dropOverLimit bool
}

// New constructs a Limiter middleware with a specified maximum number of concurrent requests, a handler to wrap, and a
// flag whether it should drop requests that exceed the limit number.
func New(maxReqs int, handler http.Handler, dropOverLimit bool) *Limiter {
	return &Limiter{
		reqs:          make(chan struct{}, maxReqs),
		handler:       handler,
		dropOverLimit: dropOverLimit,
	}
}

// ServeHTTP checks each request to determine whether it exceeds the maximum concurrent requests. It has two options
// for handling this: drop the requests and return the 503 status code, or wait for a different request to finish and
// continue the execution.
func (l *Limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.dropOverLimit {
		// if the Limiter is configured to drop limit-exceeding requests, try to send an empty struct to the channel,
		// and if it is full, return the 503 status code
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
	<-l.reqs // decrement the number of empty structs in the channel to signal the end of the request
}
