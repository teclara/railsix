package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const realtimeFreshnessThreshold = 3 * time.Minute

var realtimeHealthKeys = map[string]string{
	"alerts":          keyAlertsUpdatedAt,
	"tripUpdates":     keyTripUpdatesUpdatedAt,
	"serviceGlance":   keyServiceGlanceUpdatedAt,
	"exceptions":      keyExceptionsUpdatedAt,
	"unionDepartures": keyUnionDeparturesUpdatedAt,
}

type healthCheck struct {
	Status            string `json:"status"`
	Message           string `json:"message,omitempty"`
	AgeSeconds        int64  `json:"ageSeconds,omitempty"`
	StaleAfterSeconds int64  `json:"staleAfterSeconds,omitempty"`
}

type healthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]healthCheck `json:"checks"`
}

type staticHealthChecker interface {
	Ready(context.Context) error
}

type redisHealthChecker interface {
	Ping(context.Context) error
	GetAge(context.Context, string) (time.Duration, error)
}

func handleLiveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSONStatus(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
}

func handleReady(sc staticHealthChecker, rc redisHealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()

		body := evaluateDeparturesHealth(ctx, sc, rc)
		status := http.StatusOK
		if body.Status != "ok" {
			status = http.StatusServiceUnavailable
		}
		respondJSONStatus(w, status, body)
	}
}

func evaluateDeparturesHealth(ctx context.Context, sc staticHealthChecker, rc redisHealthChecker) healthResponse {
	checks := make(map[string]healthCheck, len(realtimeHealthKeys)+2)
	healthy := true

	if err := rc.Ping(ctx); err != nil {
		checks["redis"] = healthCheck{
			Status:  "error",
			Message: err.Error(),
		}
		healthy = false
	} else {
		checks["redis"] = healthCheck{Status: "ok"}
	}

	if err := sc.Ready(ctx); err != nil {
		checks["gtfsStatic"] = healthCheck{
			Status:  "error",
			Message: err.Error(),
		}
		healthy = false
	} else {
		checks["gtfsStatic"] = healthCheck{Status: "ok"}
	}

	for name, key := range realtimeHealthKeys {
		if checks["redis"].Status != "ok" {
			checks[name] = healthCheck{
				Status:  "error",
				Message: "redis unavailable",
			}
			continue
		}

		age, err := rc.GetAge(ctx, key)
		switch {
		case err != nil:
			checks[name] = healthCheck{
				Status:  "error",
				Message: err.Error(),
			}
			healthy = false
		case age > realtimeFreshnessThreshold:
			checks[name] = healthCheck{
				Status:            "stale",
				Message:           fmt.Sprintf("last updated %ds ago", int64(age.Seconds())),
				AgeSeconds:        int64(age.Seconds()),
				StaleAfterSeconds: int64(realtimeFreshnessThreshold.Seconds()),
			}
			healthy = false
		default:
			checks[name] = healthCheck{
				Status:            "ok",
				AgeSeconds:        int64(age.Seconds()),
				StaleAfterSeconds: int64(realtimeFreshnessThreshold.Seconds()),
			}
		}
	}

	status := "ok"
	if !healthy {
		status = "error"
	}

	return healthResponse{
		Status: status,
		Checks: checks,
	}
}
