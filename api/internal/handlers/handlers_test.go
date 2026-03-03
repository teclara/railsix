// api/internal/handlers/handlers_test.go
package handlers_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/teclara/gopulse/api/internal/cache"
	"github.com/teclara/gopulse/api/internal/handlers"
)

type mockFetcher struct {
	response []byte
	err      error
}

func (m *mockFetcher) Fetch(ctx context.Context, path string) ([]byte, error) {
	return m.response, m.err
}

func TestHealthHandler(t *testing.T) {
	h := handlers.New(nil, nil)
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Fatalf("expected ok, got %s", body["status"])
	}
}

func TestCachedProxy_CacheHit(t *testing.T) {
	c := cache.New()
	c.Set("/ServiceUpdate/UnionDepartures/All", []byte(`{"departures":[]}`), 30*time.Second)

	fetcher := &mockFetcher{response: []byte(`should not be called`)}
	h := handlers.New(fetcher, c)

	req := httptest.NewRequest("GET", "/api/departures/union", nil)
	w := httptest.NewRecorder()
	h.UnionDepartures(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"departures":[]}` {
		t.Fatalf("expected cached data, got %s", w.Body.String())
	}
}

func TestCachedProxy_CacheMiss(t *testing.T) {
	c := cache.New()
	fetcher := &mockFetcher{response: []byte(`{"departures":[{"trip":"123"}]}`)}
	h := handlers.New(fetcher, c)

	req := httptest.NewRequest("GET", "/api/departures/union", nil)
	w := httptest.NewRecorder()
	h.UnionDepartures(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"departures":[{"trip":"123"}]}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}

	// verify it was cached
	val, ok := c.Get("/ServiceUpdate/UnionDepartures/All")
	if !ok {
		t.Fatal("expected value to be cached")
	}
	if string(val) != `{"departures":[{"trip":"123"}]}` {
		t.Fatalf("unexpected cached value: %s", string(val))
	}
}
