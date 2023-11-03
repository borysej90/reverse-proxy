package limiter

import "net/http"

var _ http.Handler = (*Limiter)(nil)

type Limiter struct {
	conns         chan struct{}
	handler       http.Handler
	dropOverLimit bool
}

func New(maxConns int, handler http.Handler, dropOverLimit bool) *Limiter {
	return &Limiter{
		conns:         make(chan struct{}, maxConns),
		handler:       handler,
		dropOverLimit: dropOverLimit,
	}
}

func (l *Limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.dropOverLimit {
		select {
		case l.conns <- struct{}{}:
		default:
			http.Error(w, "server is overloaded", http.StatusServiceUnavailable)
			return
		}
	} else {
		l.conns <- struct{}{}
	}
	l.handler.ServeHTTP(w, r)
	<-l.conns
}
