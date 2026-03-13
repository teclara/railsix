package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeStaticHealth struct {
	err error
}

func (f fakeStaticHealth) Ready(context.Context) error {
	return f.err
}

type fakeRedisHealth struct {
	pingErr error
	ages    map[string]time.Duration
	ageErrs map[string]error
}

func (f fakeRedisHealth) Ping(context.Context) error {
	return f.pingErr
}

func (f fakeRedisHealth) GetAge(_ context.Context, key string) (time.Duration, error) {
	if err, ok := f.ageErrs[key]; ok {
		return 0, err
	}
	age, ok := f.ages[key]
	if !ok {
		return 0, errors.New("missing timestamp")
	}
	return age, nil
}

func TestEvaluateDeparturesHealthOK(t *testing.T) {
	ages := make(map[string]time.Duration, len(realtimeHealthKeys))
	for _, key := range realtimeHealthKeys {
		ages[key] = 30 * time.Second
	}

	result := evaluateDeparturesHealth(context.Background(), fakeStaticHealth{}, fakeRedisHealth{
		ages: ages,
	})

	if result.Status != "ok" {
		t.Fatalf("got status %q, want ok", result.Status)
	}
	if result.Checks["redis"].Status != "ok" {
		t.Fatalf("redis status = %q, want ok", result.Checks["redis"].Status)
	}
	if result.Checks["gtfsStatic"].Status != "ok" {
		t.Fatalf("gtfsStatic status = %q, want ok", result.Checks["gtfsStatic"].Status)
	}
}

func TestEvaluateDeparturesHealthStaleRealtimeData(t *testing.T) {
	ages := make(map[string]time.Duration, len(realtimeHealthKeys))
	for _, key := range realtimeHealthKeys {
		ages[key] = 30 * time.Second
	}
	ages["transit:trip-updates:updated-at"] = 5 * time.Minute

	result := evaluateDeparturesHealth(context.Background(), fakeStaticHealth{}, fakeRedisHealth{
		ages: ages,
	})

	if result.Status != "error" {
		t.Fatalf("got status %q, want error", result.Status)
	}
	if got := result.Checks["tripUpdates"].Status; got != "stale" {
		t.Fatalf("tripUpdates status = %q, want stale", got)
	}
}

func TestEvaluateDeparturesHealthDependencyFailure(t *testing.T) {
	result := evaluateDeparturesHealth(context.Background(), fakeStaticHealth{
		err: errors.New("gtfs-static unavailable"),
	}, fakeRedisHealth{
		pingErr: errors.New("redis down"),
	})

	if result.Status != "error" {
		t.Fatalf("got status %q, want error", result.Status)
	}
	if got := result.Checks["redis"].Status; got != "error" {
		t.Fatalf("redis status = %q, want error", got)
	}
	if got := result.Checks["gtfsStatic"].Status; got != "error" {
		t.Fatalf("gtfsStatic status = %q, want error", got)
	}
}

func TestHandleLivenessReturnsMinimalOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handleLiveness().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(body) != 1 || body["status"] != "ok" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestHandleReadyReturnsDependencyChecks(t *testing.T) {
	ages := make(map[string]time.Duration, len(realtimeHealthKeys))
	for _, key := range realtimeHealthKeys {
		ages[key] = 30 * time.Second
	}

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handleReady(fakeStaticHealth{}, fakeRedisHealth{
		ages: ages,
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Checks["redis"].Status != "ok" {
		t.Fatalf("redis status = %q, want ok", body.Checks["redis"].Status)
	}
}
