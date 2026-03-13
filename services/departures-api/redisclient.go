package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/teclara/railsix/shared/cache"
	"github.com/teclara/railsix/shared/gtfsrt"
	"github.com/teclara/railsix/shared/models"
)

// Redis key constants matching the realtime-poller's storage layout.
const (
	keyAlerts                   = "transit:alerts"
	keyAlertsUpdatedAt          = "transit:alerts:updated-at"
	keyTripUpdates              = "transit:trip-updates"
	keyTripUpdatesUpdatedAt     = "transit:trip-updates:updated-at"
	keyServiceGlance            = "transit:service-glance"
	keyServiceGlanceUpdatedAt   = "transit:service-glance:updated-at"
	keyExceptions               = "transit:exceptions"
	keyExceptionsUpdatedAt      = "transit:exceptions:updated-at"
	keyUnionDepartures          = "transit:union-departures"
	keyUnionDeparturesUpdatedAt = "transit:union-departures:updated-at"
)

// RedisClient wraps Redis reads needed by departure logic.
type RedisClient struct {
	rc *redis.Client
}

// NewRedisClient creates a RedisClient.
func NewRedisClient(rc *redis.Client) *RedisClient {
	return &RedisClient{rc: rc}
}

// Ping verifies Redis connectivity.
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.rc.Ping(ctx).Err()
}

// GetAge returns how old the timestamp at key is.
func (r *RedisClient) GetAge(ctx context.Context, key string) (time.Duration, error) {
	return cache.GetAge(ctx, r.rc, key)
}

// RequireFresh ensures a dataset has a recent timestamp before serving data from it.
func (r *RedisClient) RequireFresh(ctx context.Context, updatedAtKey, name string, maxAge time.Duration) error {
	age, err := r.GetAge(ctx, updatedAtKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("%s cache is not ready", name)
		}
		return fmt.Errorf("%s freshness check failed: %w", name, err)
	}
	if age > maxAge {
		return fmt.Errorf("%s cache is stale (%s old)", name, age.Round(time.Second))
	}
	return nil
}

// GetAllTripUpdates returns cached trip updates indexed by both full trip ID and trip number.
func (r *RedisClient) GetAllTripUpdates(ctx context.Context) (map[string]gtfsrt.RawTripUpdate, error) {
	updates, err := cache.GetHashAllJSON[gtfsrt.RawTripUpdate](ctx, r.rc, keyTripUpdates)
	if err != nil {
		return nil, fmt.Errorf("read trip updates: %w", err)
	}
	return updates, nil
}

// GetAllExceptions returns cancelled stop IDs indexed by trip number.
func (r *RedisClient) GetAllExceptions(ctx context.Context) (map[string][]string, error) {
	exceptions, err := cache.GetHashAllJSON[[]string](ctx, r.rc, keyExceptions)
	if err != nil {
		return nil, fmt.Errorf("read exceptions: %w", err)
	}
	return exceptions, nil
}

// GetUnionDepartures returns all cached Union Station departures.
func (r *RedisClient) GetUnionDepartures(ctx context.Context) ([]models.UnionDeparture, error) {
	var deps []models.UnionDeparture
	if err := cache.GetJSON(ctx, r.rc, keyUnionDepartures, &deps); err != nil {
		return nil, fmt.Errorf("read union departures: %w", err)
	}
	return deps, nil
}

// GetAlerts returns all cached alerts.
func (r *RedisClient) GetAlerts(ctx context.Context) ([]models.Alert, error) {
	var alerts []models.Alert
	if err := cache.GetJSON(ctx, r.rc, keyAlerts, &alerts); err != nil {
		return nil, fmt.Errorf("read alerts: %w", err)
	}
	return alerts, nil
}

// GetAllServiceGlanceMap returns all service glance entries indexed by trip number.
func (r *RedisClient) GetAllServiceGlanceMap(ctx context.Context) (map[string]models.ServiceGlanceEntry, error) {
	all, err := cache.GetHashAllJSON[models.ServiceGlanceEntry](ctx, r.rc, keyServiceGlance)
	if err != nil {
		return nil, fmt.Errorf("read service glance: %w", err)
	}
	return all, nil
}

// GetAllServiceGlance returns all service glance entries.
func (r *RedisClient) GetAllServiceGlance(ctx context.Context) ([]models.ServiceGlanceEntry, error) {
	all, err := r.GetAllServiceGlanceMap(ctx)
	if err != nil {
		return nil, err
	}
	entries := make([]models.ServiceGlanceEntry, 0, len(all))
	for _, e := range all {
		entries = append(entries, e)
	}
	return entries, nil
}

// GetNextService retrieves cached NextService data for a stop code.
func (r *RedisClient) GetNextService(ctx context.Context, stopCode string) ([]models.NextServiceLine, bool) {
	key := "transit:next-service:" + stopCode
	var lines []models.NextServiceLine
	if err := cache.GetJSON(ctx, r.rc, key, &lines); err != nil {
		return nil, false
	}
	return lines, true
}

// SetNextService caches NextService data for a stop code with 30s TTL.
func (r *RedisClient) SetNextService(ctx context.Context, stopCode string, lines []models.NextServiceLine) {
	key := "transit:next-service:" + stopCode
	_ = cache.SetJSON(ctx, r.rc, key, lines, 30*time.Second)
}
