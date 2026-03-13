package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/teclara/railsix/shared/models"
)

// httpRouteLookup resolves route IDs by calling the gtfs-static service over HTTP.
// Routes are cached in-memory since they only change on GTFS refresh (every 24h).
type httpRouteLookup struct {
	baseURL     string
	client      *http.Client
	mu          sync.RWMutex
	cache       map[string]models.Route
	gtfsVersion string
}

func newHTTPRouteLookup(baseURL string) *httpRouteLookup {
	return &httpRouteLookup{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache: make(map[string]models.Route),
	}
}

func (l *httpRouteLookup) Ready(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.baseURL+"/ready", nil)
	if err != nil {
		return fmt.Errorf("build gtfs-static readiness request: %w", err)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("gtfs-static readiness request failed: %w", err)
	}
	defer resp.Body.Close()
	l.refreshVersion(resp.Header.Get("X-GTFS-Version"))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gtfs-static readiness returned %d", resp.StatusCode)
	}

	return nil
}

// GetRoute fetches a route by ID, returning a cached result if available.
func (l *httpRouteLookup) GetRoute(id string) (models.Route, bool) {
	l.mu.RLock()
	if route, ok := l.cache[id]; ok {
		l.mu.RUnlock()
		return route, true
	}
	l.mu.RUnlock()

	routeURL := fmt.Sprintf("%s/routes/%s", l.baseURL, url.PathEscape(id))
	resp, err := l.client.Get(routeURL)
	if err != nil {
		slog.Debug("route lookup request failed", "routeID", id, "error", err)
		return models.Route{}, false
	}
	defer resp.Body.Close()
	l.refreshVersion(resp.Header.Get("X-GTFS-Version"))

	if resp.StatusCode != http.StatusOK {
		return models.Route{}, false
	}

	var route models.Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		slog.Debug("route lookup decode failed", "routeID", id, "error", err)
		return models.Route{}, false
	}

	l.mu.Lock()
	l.cache[id] = route
	l.mu.Unlock()
	return route, true
}

func (l *httpRouteLookup) refreshVersion(version string) {
	if version == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.gtfsVersion == "" {
		l.gtfsVersion = version
		return
	}
	if l.gtfsVersion == version {
		return
	}

	l.gtfsVersion = version
	l.cache = make(map[string]models.Route)
}
