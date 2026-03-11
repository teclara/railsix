package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/teclara/railsix/shared/cache"
)

const pollerFreshnessThreshold = 3 * time.Minute

var pollerHealthKeys = map[string]string{
	"alerts":          "transit:alerts:updated-at",
	"tripUpdates":     "transit:trip-updates:updated-at",
	"serviceGlance":   "transit:service-glance:updated-at",
	"exceptions":      "transit:exceptions:updated-at",
	"unionDepartures": "transit:union-departures:updated-at",
}

type pollerHealthCheck struct {
	Status            string `json:"status"`
	Message           string `json:"message,omitempty"`
	AgeSeconds        int64  `json:"ageSeconds,omitempty"`
	StaleAfterSeconds int64  `json:"staleAfterSeconds,omitempty"`
}

type pollerHealthResponse struct {
	Status string                       `json:"status"`
	Checks map[string]pollerHealthCheck `json:"checks"`
}

type natsHealthChecker interface {
	IsConnected() bool
}

type lookupHealthChecker interface {
	Ready(context.Context) error
}

type redisHealthChecker interface {
	Ping(context.Context) error
	GetAge(context.Context, string) (time.Duration, error)
}

type redisReadiness struct {
	client *redis.Client
}

func newRedisReadiness(client *redis.Client) redisHealthChecker {
	return redisReadiness{client: client}
}

func (r redisReadiness) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r redisReadiness) GetAge(ctx context.Context, key string) (time.Duration, error) {
	return cache.GetAge(ctx, r.client, key)
}

func healthHandler(nc natsHealthChecker, rc redisHealthChecker, lookup lookupHealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()

		body := evaluatePollerHealth(ctx, nc, rc, lookup)
		status := http.StatusOK
		if body.Status != "ok" {
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(body); err != nil {
			http.Error(w, "encode health response", http.StatusInternalServerError)
		}
	}
}

func evaluatePollerHealth(
	ctx context.Context,
	nc natsHealthChecker,
	rc redisHealthChecker,
	lookup lookupHealthChecker,
) pollerHealthResponse {
	checks := make(map[string]pollerHealthCheck, len(pollerHealthKeys)+3)
	healthy := true

	if nc == nil || !nc.IsConnected() {
		checks["nats"] = pollerHealthCheck{
			Status:  "error",
			Message: "nats disconnected",
		}
		healthy = false
	} else {
		checks["nats"] = pollerHealthCheck{Status: "ok"}
	}

	if err := rc.Ping(ctx); err != nil {
		checks["redis"] = pollerHealthCheck{
			Status:  "error",
			Message: err.Error(),
		}
		healthy = false
	} else {
		checks["redis"] = pollerHealthCheck{Status: "ok"}
	}

	if err := lookup.Ready(ctx); err != nil {
		checks["gtfsStatic"] = pollerHealthCheck{
			Status:  "error",
			Message: err.Error(),
		}
		healthy = false
	} else {
		checks["gtfsStatic"] = pollerHealthCheck{Status: "ok"}
	}

	for name, key := range pollerHealthKeys {
		if checks["redis"].Status != "ok" {
			checks[name] = pollerHealthCheck{
				Status:  "error",
				Message: "redis unavailable",
			}
			continue
		}

		age, err := rc.GetAge(ctx, key)
		switch {
		case err != nil:
			checks[name] = pollerHealthCheck{
				Status:  "error",
				Message: err.Error(),
			}
			healthy = false
		case age > pollerFreshnessThreshold:
			checks[name] = pollerHealthCheck{
				Status:            "stale",
				Message:           fmt.Sprintf("last updated %ds ago", int64(age.Seconds())),
				AgeSeconds:        int64(age.Seconds()),
				StaleAfterSeconds: int64(pollerFreshnessThreshold.Seconds()),
			}
			healthy = false
		default:
			checks[name] = pollerHealthCheck{
				Status:            "ok",
				AgeSeconds:        int64(age.Seconds()),
				StaleAfterSeconds: int64(pollerFreshnessThreshold.Seconds()),
			}
		}
	}

	status := "ok"
	if !healthy {
		status = "error"
	}

	return pollerHealthResponse{
		Status: status,
		Checks: checks,
	}
}
