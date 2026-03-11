package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeNATSHealth struct {
	connected bool
}

func (f fakeNATSHealth) IsConnected() bool {
	return f.connected
}

type fakeLookupHealth struct {
	err error
}

func (f fakeLookupHealth) Ready(context.Context) error {
	return f.err
}

type fakePollerRedisHealth struct {
	pingErr error
	ages    map[string]time.Duration
	ageErrs map[string]error
}

func (f fakePollerRedisHealth) Ping(context.Context) error {
	return f.pingErr
}

func (f fakePollerRedisHealth) GetAge(_ context.Context, key string) (time.Duration, error) {
	if err, ok := f.ageErrs[key]; ok {
		return 0, err
	}
	age, ok := f.ages[key]
	if !ok {
		return 0, errors.New("missing timestamp")
	}
	return age, nil
}

func TestEvaluatePollerHealthOK(t *testing.T) {
	ages := make(map[string]time.Duration, len(pollerHealthKeys))
	for _, key := range pollerHealthKeys {
		ages[key] = 30 * time.Second
	}

	result := evaluatePollerHealth(context.Background(), fakeNATSHealth{
		connected: true,
	}, fakePollerRedisHealth{
		ages: ages,
	}, fakeLookupHealth{})

	if result.Status != "ok" {
		t.Fatalf("got status %q, want ok", result.Status)
	}
	if got := result.Checks["nats"].Status; got != "ok" {
		t.Fatalf("nats status = %q, want ok", got)
	}
}

func TestEvaluatePollerHealthStaleData(t *testing.T) {
	ages := make(map[string]time.Duration, len(pollerHealthKeys))
	for _, key := range pollerHealthKeys {
		ages[key] = 30 * time.Second
	}
	ages["transit:alerts:updated-at"] = 5 * time.Minute

	result := evaluatePollerHealth(context.Background(), fakeNATSHealth{
		connected: true,
	}, fakePollerRedisHealth{
		ages: ages,
	}, fakeLookupHealth{})

	if result.Status != "error" {
		t.Fatalf("got status %q, want error", result.Status)
	}
	if got := result.Checks["alerts"].Status; got != "stale" {
		t.Fatalf("alerts status = %q, want stale", got)
	}
}

func TestEvaluatePollerHealthDependencyFailure(t *testing.T) {
	result := evaluatePollerHealth(context.Background(), fakeNATSHealth{}, fakePollerRedisHealth{
		pingErr: errors.New("redis down"),
	}, fakeLookupHealth{
		err: errors.New("gtfs-static unavailable"),
	})

	if result.Status != "error" {
		t.Fatalf("got status %q, want error", result.Status)
	}
	if got := result.Checks["nats"].Status; got != "error" {
		t.Fatalf("nats status = %q, want error", got)
	}
	if got := result.Checks["redis"].Status; got != "error" {
		t.Fatalf("redis status = %q, want error", got)
	}
	if got := result.Checks["gtfsStatic"].Status; got != "error" {
		t.Fatalf("gtfsStatic status = %q, want error", got)
	}
}
