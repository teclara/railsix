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

func TestHealthHandlerOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(fakeBrokerHealth{clients: 3}, fakeNATSConnection{connected: true}).ServeHTTP(rec, req)

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
}

func TestHealthHandlerReturns503WhenNATSIsDown(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(fakeBrokerHealth{clients: 1}, fakeNATSConnection{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got status %d, want 503", rec.Code)
	}
}
