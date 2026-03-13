package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeBrokerHealth struct {
	clients int
}

func (f fakeBrokerHealth) ClientCount() int {
	return f.clients
}

type fakeNATSConnection struct {
	connected bool
}

func (f fakeNATSConnection) IsConnected() bool {
	return f.connected
}

func TestLivenessHandlerOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	livenessHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := body["status"]; got != "ok" {
		t.Fatalf("status = %v, want ok", got)
	}
	if len(body) != 1 {
		t.Fatalf("expected minimal liveness body, got %#v", body)
	}
}

func TestReadinessHandlerReturns503WhenNATSIsDown(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	readinessHandler(fakeBrokerHealth{clients: 1}, fakeNATSConnection{}, 250).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got status %d, want 503", rec.Code)
	}
}

func TestReadinessHandlerReturns503WhenSSECapacityIsReached(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	readinessHandler(fakeBrokerHealth{clients: 250}, fakeNATSConnection{connected: true}, 250).ServeHTTP(
		rec,
		req,
	)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got status %d, want 503", rec.Code)
	}
}
