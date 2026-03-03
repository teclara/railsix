# Six Rail GTFS Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate GoPulse from proprietary Metrolinx API proxy to hybrid GTFS architecture, rename to Six Rail, and remove unused features.

**Architecture:** Open GTFS Static ZIP for reference data (stops, routes) loaded into memory on startup with daily refresh. Metrolinx API still used for GTFS-RT feeds (vehicle positions, alerts) and per-stop departures. The Go service parses and enriches data instead of proxying raw JSON.

**Tech Stack:** Go 1.22+ with `jamespfennell/gtfs` library, SvelteKit 2 with Svelte 5, Tailwind CSS 4, Mapbox GL JS.

---

### Task 1: Add models package

**Files:**
- Create: `api/internal/models/models.go`
- Test: `api/internal/models/models_test.go`

**Step 1: Write the test**

```go
// api/internal/models/models_test.go
package models_test

import (
	"encoding/json"
	"testing"

	"github.com/teclara/gopulse/api/internal/models"
)

func TestStop_JSON(t *testing.T) {
	s := models.Stop{ID: "UN", Code: "UN", Name: "Union Station", Lat: 43.6453, Lon: -79.3806}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var got models.Stop
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "Union Station" {
		t.Fatalf("expected Union Station, got %s", got.Name)
	}
}

func TestVehiclePosition_OmitsEmpty(t *testing.T) {
	vp := models.VehiclePosition{VehicleID: "V1", Lat: 43.65, Lon: -79.38}
	data, err := json.Marshal(vp)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if contains(s, "bearing") {
		t.Fatal("expected bearing to be omitted when zero")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && jsonContains(s, substr)
}

func jsonContains(s, key string) bool {
	for i := 0; i <= len(s)-len(key); i++ {
		if s[i:i+len(key)] == key {
			return true
		}
	}
	return false
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/models/ -v`
Expected: FAIL — package does not exist

**Step 3: Write implementation**

```go
// api/internal/models/models.go
package models

type Stop struct {
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	ParentID string  `json:"parentId,omitempty"`
}

type Route struct {
	ID        string `json:"id"`
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
	Color     string `json:"color"`
	TextColor string `json:"textColor"`
	Type      int    `json:"type"`
}

type VehiclePosition struct {
	VehicleID  string  `json:"vehicleId"`
	TripID     string  `json:"tripId"`
	RouteID    string  `json:"routeId"`
	RouteName  string  `json:"routeName"`
	RouteColor string  `json:"routeColor"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Bearing    float32 `json:"bearing,omitempty"`
	Speed      float32 `json:"speed,omitempty"`
	Timestamp  int64   `json:"timestamp"`
}

type Alert struct {
	ID          string   `json:"id"`
	Effect      string   `json:"effect"`
	Headline    string   `json:"headline"`
	Description string   `json:"description"`
	URL         string   `json:"url,omitempty"`
	RouteIDs    []string `json:"routeIds,omitempty"`
	RouteNames  []string `json:"routeNames,omitempty"`
	StartTime   int64    `json:"startTime,omitempty"`
	EndTime     int64    `json:"endTime,omitempty"`
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/models/ -v`
Expected: PASS

**Step 5: Commit**

```
git add api/internal/models/
git commit -m "feat(api): add typed models for stops, routes, positions, alerts"
```

---

### Task 2: Add GTFS dependency and update config

**Files:**
- Modify: `api/go.mod`
- Modify: `api/internal/config/config.go`
- Modify: `api/internal/config/config_test.go`
- Modify: `api/.env.example`

**Step 1: Add GTFS library dependency**

Run: `cd api && go get github.com/jamespfennell/gtfs@latest`

**Step 2: Update config with GTFSStaticURL**

In `api/internal/config/config.go`, add `GTFSStaticURL` field:

```go
type Config struct {
	Port             string
	MetrolinxAPIKey  string
	MetrolinxBaseURL string
	AllowedOrigins   string
	GTFSStaticURL    string
}

func Load() Config {
	return Config{
		Port:             envOr("PORT", "8080"),
		MetrolinxAPIKey:  os.Getenv("METROLINX_API_KEY"),
		MetrolinxBaseURL: envOr("METROLINX_BASE_URL", "https://api.openmetrolinx.com/OpenDataAPI/api/V1"),
		AllowedOrigins:   envOr("ALLOWED_ORIGINS", "http://localhost:5173"),
		GTFSStaticURL:    envOr("GTFS_STATIC_URL", "https://assets.metrolinx.com/raw/upload/Documents/Metrolinx/Open%20Data/GO-GTFS.zip"),
	}
}
```

**Step 3: Update .env.example**

Add `GTFS_STATIC_URL` line to `api/.env.example`:

```
METROLINX_API_KEY=your_api_key_here
GTFS_STATIC_URL=https://assets.metrolinx.com/raw/upload/Documents/Metrolinx/Open%20Data/GO-GTFS.zip
PORT=8080
ALLOWED_ORIGINS=http://localhost:5173
```

**Step 4: Run existing config tests**

Run: `cd api && go test ./internal/config/ -v`
Expected: PASS

**Step 5: Commit**

```
git add api/go.mod api/go.sum api/internal/config/ api/.env.example
git commit -m "feat(api): add jamespfennell/gtfs dependency and GTFSStaticURL config"
```

---

### Task 3: Build GTFS static store

**Files:**
- Create: `api/internal/gtfs/static.go`
- Test: `api/internal/gtfs/static_test.go`
- Create: `api/internal/gtfs/testdata/` (test fixtures)

The static store downloads the GTFS ZIP, parses it with `jamespfennell/gtfs`, and builds in-memory lookup maps. It exposes thread-safe read access and a `Refresh()` method for daily reload.

**Step 1: Write the test**

```go
// api/internal/gtfs/static_test.go
package gtfs_test

import (
	"archive/zip"
	"bytes"
	"testing"

	gtfsstore "github.com/teclara/gopulse/api/internal/gtfs"
)

// buildTestZip creates a minimal GTFS zip with stops.txt, routes.txt,
// agency.txt, calendar.txt, and trips.txt (required by parser).
func buildTestZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	files := map[string]string{
		"agency.txt":   "agency_id,agency_name,agency_url,agency_timezone\nMX,Metrolinx,https://metrolinx.com,America/Toronto\n",
		"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nWD,1,1,1,1,1,0,0,20260101,20261231\n",
		"routes.txt":   "route_id,agency_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n01,MX,LW,Lakeshore West,2,098137,FFFFFF\n09,MX,LE,Lakeshore East,2,098137,FFFFFF\n",
		"stops.txt":    "stop_id,stop_code,stop_name,stop_lat,stop_lon,location_type,parent_station\nUN,UN,Union Station,43.6453,-79.3806,1,\nUNp1,UNp1,Union Station Platform 1,43.6454,-79.3807,0,UN\n",
		"trips.txt":    "route_id,service_id,trip_id,direction_id\n01,WD,T001,0\n",
	}

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(content))
	}
	w.Close()
	return buf.Bytes()
}

func TestStaticStore_LoadFromBytes(t *testing.T) {
	zipData := buildTestZip(t)
	store, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		t.Fatal(err)
	}

	stops := store.AllStops()
	if len(stops) == 0 {
		t.Fatal("expected stops")
	}

	// Find Union Station (parent station, location_type=1)
	found := false
	for _, s := range stops {
		if s.Code == "UN" {
			found = true
			if s.Name != "Union Station" {
				t.Fatalf("expected Union Station, got %s", s.Name)
			}
		}
	}
	if !found {
		t.Fatal("Union Station not found")
	}
}

func TestStaticStore_GetRoute(t *testing.T) {
	zipData := buildTestZip(t)
	store, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		t.Fatal(err)
	}

	r, ok := store.GetRoute("01")
	if !ok {
		t.Fatal("expected route 01")
	}
	if r.ShortName != "LW" {
		t.Fatalf("expected LW, got %s", r.ShortName)
	}
	if r.Color != "098137" {
		t.Fatalf("expected 098137, got %s", r.Color)
	}
}

func TestStaticStore_GetRoute_NotFound(t *testing.T) {
	zipData := buildTestZip(t)
	store, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := store.GetRoute("99")
	if ok {
		t.Fatal("expected route 99 to not be found")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/gtfs/ -v`
Expected: FAIL — package does not exist

**Step 3: Write implementation**

```go
// api/internal/gtfs/static.go
package gtfs

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jamespfennell/gtfs"

	"github.com/teclara/gopulse/api/internal/models"
)

type StaticStore struct {
	mu     sync.RWMutex
	stops  []models.Stop
	routes map[string]models.Route
}

func NewStaticStore(zipData []byte) (*StaticStore, error) {
	s := &StaticStore{}
	if err := s.load(zipData); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StaticStore) load(zipData []byte) error {
	_ = bytes.NewReader(zipData) // validate readable
	static, err := gtfs.ParseStatic(zipData, gtfs.ParseStaticOptions{})
	if err != nil {
		return fmt.Errorf("parsing GTFS static: %w", err)
	}

	stops := make([]models.Stop, 0, len(static.Stops))
	for i := range static.Stops {
		gs := &static.Stops[i]
		if gs.Latitude == nil || gs.Longitude == nil {
			continue
		}
		var parentID string
		if gs.Parent != nil {
			parentID = gs.Parent.Id
		}
		stops = append(stops, models.Stop{
			ID:       gs.Id,
			Code:     gs.Code,
			Name:     gs.Name,
			Lat:      *gs.Latitude,
			Lon:      *gs.Longitude,
			ParentID: parentID,
		})
	}

	routes := make(map[string]models.Route, len(static.Routes))
	for i := range static.Routes {
		gr := &static.Routes[i]
		routes[gr.Id] = models.Route{
			ID:        gr.Id,
			ShortName: gr.ShortName,
			LongName:  gr.LongName,
			Color:     gr.Color,
			TextColor: gr.TextColor,
			Type:      int(gr.Type),
		}
	}

	s.mu.Lock()
	s.stops = stops
	s.routes = routes
	s.mu.Unlock()

	slog.Info("GTFS static loaded", "stops", len(stops), "routes", len(routes))
	return nil
}

func (s *StaticStore) Refresh(zipData []byte) error {
	return s.load(zipData)
}

func (s *StaticStore) AllStops() []models.Stop {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Stop, len(s.stops))
	copy(out, s.stops)
	return out
}

func (s *StaticStore) GetRoute(id string) (models.Route, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.routes[id]
	return r, ok
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/gtfs/ -v`
Expected: PASS

**Step 5: Commit**

```
git add api/internal/gtfs/
git commit -m "feat(api): add GTFS static store with stop/route parsing"
```

---

### Task 4: Build GTFS-RT realtime poller

**Files:**
- Create: `api/internal/gtfs/realtime.go`
- Test: `api/internal/gtfs/realtime_test.go`

The realtime module fetches GTFS-RT protobuf feeds from the Metrolinx API, parses them with `jamespfennell/gtfs`, enriches vehicle positions and alerts with route names from the static store, and caches the latest results in memory.

**Step 1: Write the test**

```go
// api/internal/gtfs/realtime_test.go
package gtfs_test

import (
	"testing"
	"time"

	gtfsstore "github.com/teclara/gopulse/api/internal/gtfs"
	"github.com/teclara/gopulse/api/internal/models"
)

type mockStaticLookup struct{}

func (m *mockStaticLookup) GetRoute(id string) (models.Route, bool) {
	if id == "01" {
		return models.Route{ID: "01", ShortName: "LW", LongName: "Lakeshore West", Color: "098137"}, true
	}
	return models.Route{}, false
}

func TestEnrichPositions(t *testing.T) {
	lookup := &mockStaticLookup{}
	positions := []gtfsstore.RawPosition{
		{VehicleID: "V1", TripID: "T001", RouteID: "01", Lat: 43.65, Lon: -79.38, Bearing: 180, Speed: 50, Timestamp: time.Now().Unix()},
		{VehicleID: "V2", TripID: "T002", RouteID: "99", Lat: 43.66, Lon: -79.39, Timestamp: time.Now().Unix()},
	}

	enriched := gtfsstore.EnrichPositions(positions, lookup)
	if len(enriched) != 2 {
		t.Fatalf("expected 2, got %d", len(enriched))
	}
	if enriched[0].RouteName != "Lakeshore West" {
		t.Fatalf("expected Lakeshore West, got %s", enriched[0].RouteName)
	}
	if enriched[0].RouteColor != "098137" {
		t.Fatalf("expected 098137, got %s", enriched[0].RouteColor)
	}
	// Unknown route should have empty name
	if enriched[1].RouteName != "" {
		t.Fatalf("expected empty route name for unknown route, got %s", enriched[1].RouteName)
	}
}

func TestEnrichAlerts(t *testing.T) {
	lookup := &mockStaticLookup{}
	raw := []gtfsstore.RawAlert{
		{ID: "A1", Effect: "REDUCED_SERVICE", Headline: "Delays on LW", Description: "Expect delays", RouteIDs: []string{"01"}, StartTime: time.Now().Unix()},
	}

	enriched := gtfsstore.EnrichAlerts(raw, lookup)
	if len(enriched) != 1 {
		t.Fatalf("expected 1, got %d", len(enriched))
	}
	if enriched[0].RouteNames[0] != "Lakeshore West" {
		t.Fatalf("expected Lakeshore West, got %s", enriched[0].RouteNames[0])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/gtfs/ -v -run TestEnrich`
Expected: FAIL — undefined types

**Step 3: Write implementation**

```go
// api/internal/gtfs/realtime.go
package gtfs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jamespfennell/gtfs"

	"github.com/teclara/gopulse/api/internal/models"
)

// RouteLookup is satisfied by StaticStore.
type RouteLookup interface {
	GetRoute(id string) (models.Route, bool)
}

// RawPosition holds parsed GTFS-RT vehicle data before enrichment.
type RawPosition struct {
	VehicleID string
	TripID    string
	RouteID   string
	Lat       float64
	Lon       float64
	Bearing   float32
	Speed     float32
	Timestamp int64
}

// RawAlert holds parsed GTFS-RT alert data before enrichment.
type RawAlert struct {
	ID          string
	Effect      string
	Headline    string
	Description string
	URL         string
	RouteIDs    []string
	StartTime   int64
	EndTime     int64
}

// Fetcher fetches raw bytes from a URL path (satisfied by metrolinx.Client).
type Fetcher interface {
	Fetch(ctx context.Context, path string) ([]byte, error)
}

// RealtimeCache stores the latest enriched positions and alerts.
type RealtimeCache struct {
	mu        sync.RWMutex
	positions []models.VehiclePosition
	alerts    []models.Alert
}

func NewRealtimeCache() *RealtimeCache {
	return &RealtimeCache{}
}

func (rc *RealtimeCache) SetPositions(p []models.VehiclePosition) {
	rc.mu.Lock()
	rc.positions = p
	rc.mu.Unlock()
}

func (rc *RealtimeCache) SetAlerts(a []models.Alert) {
	rc.mu.Lock()
	rc.alerts = a
	rc.mu.Unlock()
}

func (rc *RealtimeCache) GetPositions() []models.VehiclePosition {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	out := make([]models.VehiclePosition, len(rc.positions))
	copy(out, rc.positions)
	return out
}

func (rc *RealtimeCache) GetAlerts() []models.Alert {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	out := make([]models.Alert, len(rc.alerts))
	copy(out, rc.alerts)
	return out
}

// ParsePositions parses a GTFS-RT VehiclePosition protobuf feed.
func ParsePositions(data []byte) ([]RawPosition, error) {
	rt, err := gtfs.ParseRealtime(data, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing GTFS-RT positions: %w", err)
	}
	positions := make([]RawPosition, 0, len(rt.Vehicles))
	for i := range rt.Vehicles {
		v := &rt.Vehicles[i]
		if v.Position == nil || v.Position.Latitude == nil || v.Position.Longitude == nil {
			continue
		}
		var bearing float32
		if v.Position.Bearing != nil {
			bearing = *v.Position.Bearing
		}
		var speed float32
		if v.Position.Speed != nil {
			speed = *v.Position.Speed
		}
		var ts int64
		if v.Timestamp != nil {
			ts = v.Timestamp.Unix()
		}
		trip := v.GetTrip()
		positions = append(positions, RawPosition{
			VehicleID: v.GetID().ID,
			TripID:    trip.ID.ID,
			RouteID:   trip.ID.RouteID,
			Lat:       float64(*v.Position.Latitude),
			Lon:       float64(*v.Position.Longitude),
			Bearing:   bearing,
			Speed:     speed,
			Timestamp: ts,
		})
	}
	return positions, nil
}

// ParseAlerts parses a GTFS-RT Alerts protobuf feed.
func ParseAlerts(data []byte) ([]RawAlert, error) {
	rt, err := gtfs.ParseRealtime(data, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing GTFS-RT alerts: %w", err)
	}
	alerts := make([]RawAlert, 0, len(rt.Alerts))
	for i := range rt.Alerts {
		a := &rt.Alerts[i]
		headline := ""
		if len(a.Header) > 0 {
			headline = a.Header[0].Text
		}
		desc := ""
		if len(a.Description) > 0 {
			desc = a.Description[0].Text
		}
		url := ""
		if len(a.URL) > 0 {
			url = a.URL[0].Text
		}
		var routeIDs []string
		for _, ie := range a.InformedEntities {
			if ie.RouteID != nil {
				routeIDs = append(routeIDs, *ie.RouteID)
			}
		}
		var startTime, endTime int64
		if len(a.ActivePeriods) > 0 {
			if a.ActivePeriods[0].StartsAt != nil {
				startTime = a.ActivePeriods[0].StartsAt.Unix()
			}
			if a.ActivePeriods[0].EndsAt != nil {
				endTime = a.ActivePeriods[0].EndsAt.Unix()
			}
		}
		alerts = append(alerts, RawAlert{
			ID:          a.ID,
			Effect:      a.Effect.String(),
			Headline:    headline,
			Description: desc,
			URL:         url,
			RouteIDs:    routeIDs,
			StartTime:   startTime,
			EndTime:     endTime,
		})
	}
	return alerts, nil
}

// EnrichPositions adds route name and color from static data.
func EnrichPositions(raw []RawPosition, lookup RouteLookup) []models.VehiclePosition {
	out := make([]models.VehiclePosition, len(raw))
	for i, rp := range raw {
		vp := models.VehiclePosition{
			VehicleID: rp.VehicleID,
			TripID:    rp.TripID,
			RouteID:   rp.RouteID,
			Lat:       rp.Lat,
			Lon:       rp.Lon,
			Bearing:   rp.Bearing,
			Speed:     rp.Speed,
			Timestamp: rp.Timestamp,
		}
		if route, ok := lookup.GetRoute(rp.RouteID); ok {
			vp.RouteName = route.LongName
			vp.RouteColor = route.Color
		}
		out[i] = vp
	}
	return out
}

// EnrichAlerts adds route names from static data.
func EnrichAlerts(raw []RawAlert, lookup RouteLookup) []models.Alert {
	out := make([]models.Alert, len(raw))
	for i, ra := range raw {
		alert := models.Alert{
			ID:          ra.ID,
			Effect:      ra.Effect,
			Headline:    ra.Headline,
			Description: ra.Description,
			URL:         ra.URL,
			RouteIDs:    ra.RouteIDs,
			StartTime:   ra.StartTime,
			EndTime:     ra.EndTime,
		}
		for _, rid := range ra.RouteIDs {
			if route, ok := lookup.GetRoute(rid); ok {
				alert.RouteNames = append(alert.RouteNames, route.LongName)
			}
		}
		out[i] = alert
	}
	return out
}

// StartPositionPoller polls the GTFS-RT VehiclePosition feed on an interval.
func StartPositionPoller(ctx context.Context, fetcher Fetcher, lookup RouteLookup, cache *RealtimeCache, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	// Run once immediately
	pollPositions(ctx, fetcher, lookup, cache)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pollPositions(ctx, fetcher, lookup, cache)
		}
	}
}

func pollPositions(ctx context.Context, fetcher Fetcher, lookup RouteLookup, cache *RealtimeCache) {
	data, err := fetcher.Fetch(ctx, "/Gtfs/Feed/VehiclePosition")
	if err != nil {
		slog.Error("failed to fetch vehicle positions", "error", err)
		return
	}
	raw, err := ParsePositions(data)
	if err != nil {
		slog.Error("failed to parse vehicle positions", "error", err)
		return
	}
	enriched := EnrichPositions(raw, lookup)
	cache.SetPositions(enriched)
	slog.Debug("updated vehicle positions", "count", len(enriched))
}

// StartAlertPoller polls the GTFS-RT Alerts feed on an interval.
func StartAlertPoller(ctx context.Context, fetcher Fetcher, lookup RouteLookup, cache *RealtimeCache, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	pollAlerts(ctx, fetcher, lookup, cache)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pollAlerts(ctx, fetcher, lookup, cache)
		}
	}
}

func pollAlerts(ctx context.Context, fetcher Fetcher, lookup RouteLookup, cache *RealtimeCache) {
	data, err := fetcher.Fetch(ctx, "/Gtfs/Feed/Alerts")
	if err != nil {
		slog.Error("failed to fetch alerts", "error", err)
		return
	}
	raw, err := ParseAlerts(data)
	if err != nil {
		slog.Error("failed to parse alerts", "error", err)
		return
	}
	enriched := EnrichAlerts(raw, lookup)
	cache.SetAlerts(enriched)
	slog.Debug("updated alerts", "count", len(enriched))
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/gtfs/ -v`
Expected: PASS (all tests including static and realtime)

**Step 5: Commit**

```
git add api/internal/gtfs/realtime.go api/internal/gtfs/realtime_test.go
git commit -m "feat(api): add GTFS-RT parser, enrichment, and background pollers"
```

---

### Task 5: Rewrite handlers

**Files:**
- Modify: `api/internal/handlers/handlers.go`
- Modify: `api/internal/handlers/handlers_test.go`

Replace the proxy-all-endpoints approach with four focused handlers that serve typed JSON. Remove all deleted endpoints. Keep `cachedProxy` for departures (still hits Metrolinx REST). Positions and alerts are served from the `RealtimeCache`. Stops are served from the `StaticStore`.

**Step 1: Write the updated tests**

```go
// api/internal/handlers/handlers_test.go
package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/teclara/gopulse/api/internal/cache"
	gtfsstore "github.com/teclara/gopulse/api/internal/gtfs"
	"github.com/teclara/gopulse/api/internal/handlers"
	"github.com/teclara/gopulse/api/internal/models"
)

type mockFetcher struct {
	response []byte
	err      error
}

func (m *mockFetcher) Fetch(ctx context.Context, path string) ([]byte, error) {
	return m.response, m.err
}

func TestHealthHandler(t *testing.T) {
	h := handlers.New(nil, nil, nil, nil)
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

func TestStopsHandler(t *testing.T) {
	store := mustBuildStore(t)
	h := handlers.New(nil, nil, store, nil)
	req := httptest.NewRequest("GET", "/api/stops", nil)
	w := httptest.NewRecorder()

	h.AllStops(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var stops []models.Stop
	json.Unmarshal(w.Body.Bytes(), &stops)
	if len(stops) == 0 {
		t.Fatal("expected stops")
	}
}

func TestPositionsHandler(t *testing.T) {
	rtCache := gtfsstore.NewRealtimeCache()
	rtCache.SetPositions([]models.VehiclePosition{
		{VehicleID: "V1", TripID: "T1", RouteID: "01", RouteName: "Lakeshore West", Lat: 43.65, Lon: -79.38},
	})
	h := handlers.New(nil, nil, nil, rtCache)
	req := httptest.NewRequest("GET", "/api/positions", nil)
	w := httptest.NewRecorder()

	h.Positions(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var positions []models.VehiclePosition
	json.Unmarshal(w.Body.Bytes(), &positions)
	if len(positions) != 1 || positions[0].RouteName != "Lakeshore West" {
		t.Fatalf("unexpected positions: %+v", positions)
	}
}

func TestAlertsHandler(t *testing.T) {
	rtCache := gtfsstore.NewRealtimeCache()
	rtCache.SetAlerts([]models.Alert{
		{ID: "A1", Headline: "Delays", Effect: "REDUCED_SERVICE"},
	})
	h := handlers.New(nil, nil, nil, rtCache)
	req := httptest.NewRequest("GET", "/api/alerts", nil)
	w := httptest.NewRecorder()

	h.Alerts(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var alerts []models.Alert
	json.Unmarshal(w.Body.Bytes(), &alerts)
	if len(alerts) != 1 || alerts[0].Headline != "Delays" {
		t.Fatalf("unexpected alerts: %+v", alerts)
	}
}

func TestStopDepartures_CacheHit(t *testing.T) {
	c := cache.New()
	c.Set("/Stop/NextService/UN", []byte(`[{"LineName":"LW"}]`), 30*time.Second)
	h := handlers.New(&mockFetcher{}, c, nil, nil)

	req := httptest.NewRequest("GET", "/api/departures/UN", nil)
	req.SetPathValue("stopCode", "UN")
	w := httptest.NewRecorder()
	h.StopDepartures(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestStopDepartures_InvalidCode(t *testing.T) {
	h := handlers.New(nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/api/departures/../etc", nil)
	req.SetPathValue("stopCode", "../etc")
	w := httptest.NewRecorder()
	h.StopDepartures(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
```

Add a helper to build a StaticStore from test data (reuse `buildTestZip` approach or inline):

```go
func mustBuildStore(t *testing.T) *gtfsstore.StaticStore {
	t.Helper()
	zipData := buildTestZip(t)
	store, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		t.Fatal(err)
	}
	return store
}
```

Note: `buildTestZip` needs to be duplicated or extracted. For simplicity, duplicate it in this test file.

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/handlers/ -v`
Expected: FAIL — `handlers.New` signature changed, new methods missing

**Step 3: Rewrite handlers**

```go
// api/internal/handlers/handlers.go
package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/teclara/gopulse/api/internal/cache"
	gtfsstore "github.com/teclara/gopulse/api/internal/gtfs"
	"github.com/teclara/gopulse/api/internal/models"
)

var stopCodeRe = regexp.MustCompile(`^[A-Za-z0-9]{2,10}$`)

type Fetcher interface {
	Fetch(ctx context.Context, path string) ([]byte, error)
}

type Handlers struct {
	fetcher Fetcher
	cache   *cache.Cache
	static  *gtfsstore.StaticStore
	rt      *gtfsstore.RealtimeCache
}

func New(fetcher Fetcher, cache *cache.Cache, static *gtfsstore.StaticStore, rt *gtfsstore.RealtimeCache) *Handlers {
	return &Handlers{fetcher: fetcher, cache: cache, static: static, rt: rt}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func writeJSON(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Warn("write response failed", "error", err)
	}
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, []byte(`{"error":"`+msg+`"}`))
}

func respondJSON(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// AllStops serves stops from GTFS static data.
func (h *Handlers) AllStops(w http.ResponseWriter, r *http.Request) {
	stops := h.static.AllStops()
	respondJSON(w, stops)
}

// Positions serves enriched vehicle positions from the realtime cache.
func (h *Handlers) Positions(w http.ResponseWriter, r *http.Request) {
	positions := h.rt.GetPositions()
	if positions == nil {
		positions = []models.VehiclePosition{}
	}
	respondJSON(w, positions)
}

// Alerts serves enriched alerts from the realtime cache.
func (h *Handlers) Alerts(w http.ResponseWriter, r *http.Request) {
	alerts := h.rt.GetAlerts()
	if alerts == nil {
		alerts = []models.Alert{}
	}
	respondJSON(w, alerts)
}

// StopDepartures still proxies via the Metrolinx REST API (no GTFS-RT equivalent).
func (h *Handlers) StopDepartures(w http.ResponseWriter, r *http.Request) {
	stopCode := r.PathValue("stopCode")
	if !stopCodeRe.MatchString(stopCode) {
		jsonError(w, "invalid stop code", http.StatusBadRequest)
		return
	}
	h.cachedProxy(w, r, "/Stop/NextService/"+stopCode, 30*time.Second)
}

func (h *Handlers) cachedProxy(w http.ResponseWriter, r *http.Request, metrolinxPath string, ttl time.Duration) {
	if data, ok := h.cache.Get(metrolinxPath); ok {
		w.Header().Set("X-Cache", "HIT")
		writeJSON(w, http.StatusOK, data)
		return
	}

	data, err := h.fetcher.Fetch(r.Context(), metrolinxPath)
	if err != nil {
		slog.Error("metrolinx fetch failed", "path", metrolinxPath, "error", err)
		if stale, ok := h.cache.GetStale(metrolinxPath); ok {
			w.Header().Set("X-Cache", "STALE")
			w.Header().Set("X-Cache-Stale", "true")
			writeJSON(w, http.StatusOK, stale)
			return
		}
		jsonError(w, "upstream unavailable", http.StatusBadGateway)
		return
	}

	h.cache.Set(metrolinxPath, data, ttl)
	w.Header().Set("X-Cache", "MISS")
	writeJSON(w, http.StatusOK, data)
}
```

**Step 4: Run tests**

Run: `cd api && go test ./internal/handlers/ -v`
Expected: PASS

**Step 5: Commit**

```
git add api/internal/handlers/
git commit -m "feat(api): rewrite handlers for GTFS-backed stops, positions, alerts"
```

---

### Task 6: Rewrite main.go with GTFS startup and pollers

**Files:**
- Modify: `api/cmd/server/main.go`

Wire up the GTFS static store (download ZIP on startup), realtime cache, background pollers, and the new handler signatures. Register only the four remaining endpoints.

**Step 1: Write updated main.go**

```go
// api/cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/teclara/gopulse/api/internal/cache"
	"github.com/teclara/gopulse/api/internal/config"
	gtfsstore "github.com/teclara/gopulse/api/internal/gtfs"
	"github.com/teclara/gopulse/api/internal/handlers"
	"github.com/teclara/gopulse/api/internal/metrolinx"
)

func main() {
	cfg := config.Load()

	// Download and parse GTFS static data
	slog.Info("downloading GTFS static data", "url", cfg.GTFSStaticURL)
	zipData, err := downloadURL(cfg.GTFSStaticURL)
	if err != nil {
		slog.Error("failed to download GTFS static data", "error", err)
		os.Exit(1)
	}
	static, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		slog.Error("failed to parse GTFS static data", "error", err)
		os.Exit(1)
	}

	// Start daily GTFS refresh
	go refreshLoop(cfg.GTFSStaticURL, static, 24*time.Hour)

	// Metrolinx client for REST departures + GTFS-RT feeds
	client := metrolinx.NewClient(cfg.MetrolinxBaseURL, cfg.MetrolinxAPIKey)

	// Realtime cache + background pollers
	rtCache := gtfsstore.NewRealtimeCache()
	ctx := context.Background()
	go gtfsstore.StartPositionPoller(ctx, client, static, rtCache, 10*time.Second)
	go gtfsstore.StartAlertPoller(ctx, client, static, rtCache, 30*time.Second)

	// Departures still use the TTL cache for proxying
	c := cache.New()
	h := handlers.New(client, c, static, rtCache)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/stops", h.AllStops)
	mux.HandleFunc("GET /api/departures/{stopCode}", h.StopDepartures)
	mux.HandleFunc("GET /api/positions", h.Positions)
	mux.HandleFunc("GET /api/alerts", h.Alerts)

	handler := corsMiddleware(cfg.AllowedOrigins, mux)

	slog.Info("starting sixrail-api", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func downloadURL(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	const maxBytes = 50 * 1024 * 1024 // 50 MB
	return io.ReadAll(io.LimitReader(resp.Body, maxBytes))
}

func refreshLoop(url string, static *gtfsstore.StaticStore, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		slog.Info("refreshing GTFS static data")
		data, err := downloadURL(url)
		if err != nil {
			slog.Error("failed to download GTFS refresh", "error", err)
			continue
		}
		if err := static.Refresh(data); err != nil {
			slog.Error("failed to parse GTFS refresh", "error", err)
		}
	}
}

func corsMiddleware(allowedOrigins string, next http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

**Step 2: Verify it compiles**

Run: `cd api && go build ./cmd/server/`
Expected: Compiles without errors

**Step 3: Run all API tests**

Run: `cd api && go test ./... -v`
Expected: All tests PASS

**Step 4: Commit**

```
git add api/cmd/server/main.go
git commit -m "feat(api): wire GTFS static store and RT pollers into server startup"
```

---

### Task 7: Update frontend API client and layout

**Files:**
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/routes/+layout.server.ts`
- Modify: `web/src/lib/components/AlertBanner.svelte`

Strip out removed endpoints, update to new API shapes.

**Step 1: Rewrite api.ts**

```typescript
// web/src/lib/api.ts
import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';

function getBaseUrl() {
	const url = env.API_BASE_URL || (dev ? 'http://localhost:8080' : '');
	if (!url) {
		throw new Error('API_BASE_URL environment variable is required in production');
	}
	return url;
}

async function fetchApi<T>(path: string): Promise<T> {
	const res = await fetch(`${getBaseUrl()}${path}`);
	if (!res.ok) {
		throw new Error(`API error: ${res.status} ${res.statusText}`);
	}
	return res.json();
}

export interface Stop {
	id: string;
	code: string;
	name: string;
	lat: number;
	lon: number;
	parentId?: string;
}

export interface VehiclePosition {
	vehicleId: string;
	tripId: string;
	routeId: string;
	routeName: string;
	routeColor: string;
	lat: number;
	lon: number;
	bearing?: number;
	speed?: number;
	timestamp: number;
}

export interface Alert {
	id: string;
	effect: string;
	headline: string;
	description: string;
	url?: string;
	routeIds?: string[];
	routeNames?: string[];
	startTime?: number;
	endTime?: number;
}

export function getAllStops() {
	return fetchApi<Stop[]>('/api/stops');
}

export function getStopDepartures(stopCode: string) {
	return fetchApi(`/api/departures/${stopCode}`);
}

export function getPositions() {
	return fetchApi<VehiclePosition[]>('/api/positions');
}

export function getAlerts() {
	return fetchApi<Alert[]>('/api/alerts');
}
```

**Step 2: Update layout server to use new `getAlerts`**

```typescript
// web/src/routes/+layout.server.ts
import { getAlerts } from '$lib/api';

export async function load() {
	try {
		const alerts = await getAlerts();
		return { alerts: Array.isArray(alerts) ? alerts : [] };
	} catch {
		return { alerts: [] };
	}
}
```

**Step 3: Update AlertBanner to use new alert shape**

```svelte
<!-- web/src/lib/components/AlertBanner.svelte -->
<script lang="ts">
	import type { Alert } from '$lib/api';
	let { alerts = [] }: { alerts: Alert[] } = $props();
</script>

{#if alerts.length > 0}
	<div class="bg-amber-100 border-b border-amber-300 px-4 py-2 text-sm text-amber-900">
		<strong>Service Alert:</strong> {alerts[0].headline || 'Service disruption reported'}
		{#if alerts.length > 1}
			<a href="/alerts" class="underline ml-2">+{alerts.length - 1} more</a>
		{/if}
	</div>
{/if}
```

**Step 4: Verify frontend compiles**

Run: `cd web && npm run check`
Expected: No TypeScript errors (some warnings may exist from unchanged files)

**Step 5: Commit**

```
git add web/src/lib/api.ts web/src/routes/+layout.server.ts web/src/lib/components/AlertBanner.svelte
git commit -m "feat(web): update API client and layout for new GTFS-backed endpoints"
```

---

### Task 8: Update map page for enriched positions

**Files:**
- Modify: `web/src/routes/map/+page.server.ts`
- Modify: `web/src/routes/map/+page.svelte`

Use the new `getPositions()` which returns a flat array of `VehiclePosition` objects (no more `entity.vehicle.position` nesting).

**Step 1: Update map server loader**

```typescript
// web/src/routes/map/+page.server.ts
import { getPositions } from '$lib/api';

export async function load() {
	try {
		const positions = await getPositions();
		return { positions: Array.isArray(positions) ? positions : [] };
	} catch {
		return { positions: [] };
	}
}
```

**Step 2: Update map page component**

```svelte
<!-- web/src/routes/map/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { invalidateAll } from '$app/navigation';
	import { env } from '$env/dynamic/public';
	import type { VehiclePosition } from '$lib/api';

	let { data } = $props();
	let mapContainer: HTMLDivElement;
	let map: any;
	let markers: any[] = [];
	let mapboxgl: any;

	function updateMarkers() {
		if (!map || !mapboxgl) return;

		markers.forEach((m) => m.remove());
		markers = [];

		for (const pos of data.positions as VehiclePosition[]) {
			if (!pos.lat || !pos.lon) continue;
			const color = pos.routeColor ? `#${pos.routeColor}` : '#15803d';
			const m = new mapboxgl.Marker({ color })
				.setLngLat([pos.lon, pos.lat])
				.setPopup(
					new mapboxgl.Popup().setHTML(
						`<strong>${pos.routeName || pos.routeId || '—'}</strong><br/>
						 Trip: ${pos.tripId || '—'}`
					)
				)
				.addTo(map);
			markers.push(m);
		}
	}

	onMount(() => {
		(async () => {
			mapboxgl = (await import('mapbox-gl')).default;
			mapboxgl.accessToken = env.PUBLIC_MAPBOX_TOKEN || '';
			map = new mapboxgl.Map({
				container: mapContainer,
				style: 'mapbox://styles/mapbox/light-v11',
				center: [-79.38, 43.65],
				zoom: 9
			});
			map.addControl(new mapboxgl.NavigationControl());
			updateMarkers();
		})();

		const interval = setInterval(() => invalidateAll(), 15_000);
		return () => {
			clearInterval(interval);
			markers.forEach((m) => m.remove());
			map?.remove();
		};
	});

	$effect(() => {
		data.positions;
		updateMarkers();
	});
</script>

<svelte:head>
	<link href="https://api.mapbox.com/mapbox-gl-js/v3.4.0/mapbox-gl.css" rel="stylesheet" />
</svelte:head>

<div class="space-y-4">
	<h1 class="text-2xl font-bold">Live Train Map</h1>
	<div bind:this={mapContainer} class="w-full h-[600px] rounded-lg border border-gray-200"></div>
</div>
```

**Step 3: Verify frontend compiles**

Run: `cd web && npm run check`
Expected: No errors

**Step 4: Commit**

```
git add web/src/routes/map/
git commit -m "feat(web): update map page for enriched GTFS-RT vehicle positions"
```

---

### Task 9: Update alerts page

**Files:**
- Modify: `web/src/routes/alerts/+page.server.ts`
- Modify: `web/src/routes/alerts/+page.svelte`

The new API returns a single `/api/alerts` array (no more separate service/info/exceptions). Simplify the page to show all alerts with their effect type as a filter.

**Step 1: Update alerts server loader**

```typescript
// web/src/routes/alerts/+page.server.ts
import { getAlerts } from '$lib/api';

export async function load() {
	try {
		const alerts = await getAlerts();
		return { alerts: Array.isArray(alerts) ? alerts : [] };
	} catch {
		return { alerts: [] };
	}
}
```

**Step 2: Update alerts page component**

```svelte
<!-- web/src/routes/alerts/+page.svelte -->
<script lang="ts">
	import type { Alert } from '$lib/api';

	let { data } = $props();
	let filter = $state('all');

	let filtered = $derived(
		filter === 'all'
			? data.alerts
			: (data.alerts as Alert[]).filter((a) => a.effect === filter)
	);

	let effects = $derived(
		[...new Set((data.alerts as Alert[]).map((a) => a.effect))].sort()
	);
</script>

<div class="space-y-4">
	<h1 class="text-2xl font-bold">Service Alerts</h1>

	<div class="flex gap-2 flex-wrap">
		<button onclick={() => filter = 'all'} class="px-3 py-1 rounded {filter === 'all' ? 'bg-green-700 text-white' : 'bg-gray-200'}">All</button>
		{#each effects as effect}
			<button onclick={() => filter = effect} class="px-3 py-1 rounded {filter === effect ? 'bg-red-600 text-white' : 'bg-gray-200'}">{effect.replace(/_/g, ' ')}</button>
		{/each}
	</div>

	{#each filtered as alert (alert.id)}
		<div class="bg-red-50 border-l-4 border-red-500 p-4 rounded">
			<p class="font-medium text-red-900">{alert.headline || 'Service disruption'}</p>
			{#if alert.description}
				<p class="text-sm text-red-800 mt-1">{alert.description}</p>
			{/if}
			{#if alert.routeNames?.length}
				<p class="text-sm text-red-700 mt-1">Routes: {alert.routeNames.join(', ')}</p>
			{/if}
			{#if alert.url}
				<a href={alert.url} target="_blank" class="text-sm text-red-700 underline mt-1 block">More info</a>
			{/if}
		</div>
	{:else}
		<p class="text-gray-500 py-8 text-center">No active alerts.</p>
	{/each}
</div>
```

**Step 3: Verify frontend compiles**

Run: `cd web && npm run check`
Expected: No errors

**Step 4: Commit**

```
git add web/src/routes/alerts/
git commit -m "feat(web): update alerts page for unified GTFS-RT alerts endpoint"
```

---

### Task 10: Update remaining frontend pages (home, departures, stations)

**Files:**
- Modify: `web/src/routes/+page.svelte`
- Modify: `web/src/routes/+page.server.ts`
- Modify: `web/src/routes/departures/[stopCode]/+page.server.ts`
- Modify: `web/src/routes/stations/+page.server.ts`
- Modify: `web/src/routes/stations/+page.svelte`
- Modify: `web/src/lib/components/StationPicker.svelte`
- Modify: `web/src/lib/components/Nav.svelte`

Stops now return `{id, code, name, lat, lon}` instead of `{StopCode, StopName, ...}`. Update field access. Departures page no longer calls `getStopDetails` (removed). Nav drops schedule/journey links.

**Step 1: Update Nav — remove schedule/journey links, rename to Six Rail**

```svelte
<!-- web/src/lib/components/Nav.svelte -->
<script lang="ts">
	import { page } from '$app/stores';

	let pathname = $derived($page.url.pathname as string);
</script>

<nav class="bg-green-700 text-white">
	<div class="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
		<a href="/" class="text-xl font-bold tracking-tight">Six Rail</a>
		<div class="flex gap-4 text-sm">
			<a href="/stations" class:font-bold={pathname === '/stations'}>Stations</a>
			<a href="/map" class:font-bold={pathname === '/map'}>Map</a>
			<a href="/alerts" class:font-bold={pathname === '/alerts'}>Alerts</a>
		</div>
	</div>
</nav>
```

**Step 2: Update home page text**

In `web/src/routes/+page.svelte`, change "GoPulse" to "Six Rail" and update tagline:

```svelte
<h1 class="text-4xl font-bold text-gray-900 mb-2">Six Rail</h1>
<p class="text-gray-600 mb-8">Real-time train tracking</p>
```

**Step 3: Update StationPicker to use new stop fields**

Read `web/src/lib/components/StationPicker.svelte` first, then update field access from `stop.StopCode`/`stop.StopName` to `stop.code`/`stop.name`.

**Step 4: Update stations page to use new stop fields**

Read `web/src/routes/stations/+page.svelte` first, then update field access from `stop.StopCode`/`stop.StopName`/`stop.Code`/`stop.Name` to `stop.code`/`stop.name`.

**Step 5: Update departures page server — remove getStopDetails call**

```typescript
// web/src/routes/departures/[stopCode]/+page.server.ts
import { getStopDepartures } from '$lib/api';

export async function load({ params }) {
	try {
		const departures = await getStopDepartures(params.stopCode);
		return {
			stopCode: params.stopCode,
			departures: Array.isArray(departures) ? departures : []
		};
	} catch {
		return { stopCode: params.stopCode, departures: [] };
	}
}
```

**Step 6: Update departures page svelte — remove stopDetails references**

In `web/src/routes/departures/[stopCode]/+page.svelte`, change the header from `data.stopDetails?.StopName` to just `Station ${data.stopCode}` (stop details endpoint was removed).

**Step 7: Verify frontend compiles**

Run: `cd web && npm run check`
Expected: No errors

**Step 8: Commit**

```
git add web/src/routes/ web/src/lib/components/
git commit -m "feat(web): update all pages for GTFS data shape, rename to Six Rail"
```

---

### Task 11: Remove dead code

**Files:**
- Delete: `web/src/routes/schedule/` (entire directory)
- Delete: `web/src/routes/journey/` (entire directory)

**Step 1: Remove schedule and journey pages**

```bash
rm -rf web/src/routes/schedule web/src/routes/journey
```

**Step 2: Verify frontend still compiles**

Run: `cd web && npm run check`
Expected: No errors

**Step 3: Run all API tests**

Run: `cd api && go test ./... -v`
Expected: All tests PASS

**Step 4: Commit**

```
git add -A
git commit -m "chore: remove schedule and journey pages"
```

---

### Task 12: Rename Go module and update package.json

**Files:**
- Modify: `api/go.mod` — change module path
- Modify: all Go files with import paths
- Modify: `web/package.json` — change name

**Step 1: Rename Go module**

In `api/go.mod`, change:
```
module github.com/teclara/gopulse/api
```
to:
```
module github.com/teclara/sixrail/api
```

**Step 2: Update all Go import paths**

Find and replace `github.com/teclara/gopulse/api` → `github.com/teclara/sixrail/api` in all `.go` files under `api/`.

Files to update:
- `api/cmd/server/main.go`
- `api/internal/handlers/handlers.go`
- `api/internal/handlers/handlers_test.go`
- `api/internal/gtfs/static_test.go`
- `api/internal/gtfs/realtime.go`
- `api/internal/gtfs/realtime_test.go` (if importing models)

**Step 3: Update web/package.json name**

Change `"name": "web"` to `"name": "sixrail-web"`.

**Step 4: Verify everything compiles**

```bash
cd api && go build ./cmd/server/ && go test ./... -v
cd ../web && npm run check
```
Expected: All pass

**Step 5: Commit**

```
git add -A
git commit -m "chore: rename module from gopulse to sixrail"
```

---

### Task 13: Update deploy config and env examples

**Files:**
- Modify: `api/railway.toml` — update watch patterns if needed
- Modify: `web/railway.toml` — update watch patterns if needed
- Modify: `api/.env.example`
- Modify: `web/.env.example`

**Step 1: Verify railway.toml files are correct**

The `api/railway.toml` watchPatterns `["api/**"]` and `web/railway.toml` watchPatterns `["web/**"]` are still correct since directory names haven't changed.

No changes needed to `railway.toml` files.

**Step 2: Verify .env.example files are up to date**

`api/.env.example` should have (from Task 2):
```
METROLINX_API_KEY=your_api_key_here
GTFS_STATIC_URL=https://assets.metrolinx.com/raw/upload/Documents/Metrolinx/Open%20Data/GO-GTFS.zip
PORT=8080
ALLOWED_ORIGINS=http://localhost:5173
```

`web/.env.example` stays the same:
```
API_BASE_URL=http://localhost:8080
PUBLIC_MAPBOX_TOKEN=your_mapbox_token_here
```

**Step 3: Run full test suite one final time**

```bash
cd api && go test ./... -v
cd ../web && npm run check && npm run build
```
Expected: All pass, build succeeds

**Step 4: Commit (if any changes)**

```
git add -A
git commit -m "chore: update deploy config and env examples for Six Rail"
```
