package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConcurrencyLimiterRejectsWhenSaturated(t *testing.T) {
	limiter := newConcurrencyLimiter(1)
	started := make(chan struct{})
	release := make(chan struct{})

	handler := limiter.wrap(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusNoContent)
	})

	firstDone := make(chan struct{})
	go func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/departures/UN", nil)
		handler.ServeHTTP(rec, req)
		close(firstDone)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first request did not acquire limiter")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/departures/UN", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if got := rec.Header().Get("Retry-After"); got != "1" {
		t.Fatalf("Retry-After = %q, want %q", got, "1")
	}

	close(release)

	select {
	case <-firstDone:
	case <-time.After(time.Second):
		t.Fatal("first request did not complete")
	}
}
