package main

import "net/http"

type concurrencyLimiter struct {
	tokens chan struct{}
}

func newConcurrencyLimiter(maxConcurrent int) *concurrencyLimiter {
	return &concurrencyLimiter{
		tokens: make(chan struct{}, maxConcurrent),
	}
}

func (l *concurrencyLimiter) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case l.tokens <- struct{}{}:
			defer func() {
				<-l.tokens
			}()
		default:
			w.Header().Set("Retry-After", "1")
			jsonError(w, "service is temporarily overloaded", http.StatusServiceUnavailable)
			return
		}

		next(w, r)
	}
}
