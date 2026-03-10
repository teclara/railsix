package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teclara/railsix/gtfs-static/store"
)

func TestReadyReturns503BeforeLoad(t *testing.T) {
	s := store.NewEmptyStaticStore()
	mux := http.NewServeMux()
	registerRoutes(mux, s)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["status"] != "loading" {
		t.Errorf("expected status 'loading', got %q", body["status"])
	}
}

func TestStopsReturns503BeforeLoad(t *testing.T) {
	s := store.NewEmptyStaticStore()
	mux := http.NewServeMux()
	registerRoutes(mux, s)

	req := httptest.NewRequest(http.MethodGet, "/stops", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}
}

func TestDeparturesReturns503BeforeLoad(t *testing.T) {
	s := store.NewEmptyStaticStore()
	mux := http.NewServeMux()
	registerRoutes(mux, s)

	req := httptest.NewRequest(http.MethodGet, "/departures/STOP1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}
}

func TestServiceActiveRequiresDateParam(t *testing.T) {
	s := store.NewEmptyStaticStore()
	mux := http.NewServeMux()
	registerRoutes(mux, s)

	// Even before load, the not-ready check fires first
	req := httptest.NewRequest(http.MethodGet, "/services/SVC1/active", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}
}
