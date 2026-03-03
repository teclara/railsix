# Six Rail — GTFS Migration Design

## Context

GoPulse currently proxies the Metrolinx proprietary API, which has restrictive terms prohibiting data redistribution through your own API (Section 7a). This migration switches to a hybrid approach: open GTFS Static data for reference, proprietary API only where no open alternative exists (real-time feeds, per-stop departures). The project is also renamed from GoPulse to Six Rail to avoid GO Transit trademark concerns.

## Data Sources

| Source | URL | License | Data |
|---|---|---|---|
| GTFS Static ZIP | `https://assets.metrolinx.com/raw/upload/Documents/Metrolinx/Open%20Data/GO-GTFS.zip` | Open Government Licence – Ontario | Stops, routes, trips, stop_times, shapes |
| Metrolinx API (GTFS-RT) | `api.openmetrolinx.com/OpenDataAPI/api/V1/Gtfs/Feed/*` | Proprietary (API key required) | Vehicle positions, alerts |
| Metrolinx API (REST) | `api.openmetrolinx.com/OpenDataAPI/api/V1/Stop/NextService/*` | Proprietary (API key required) | Per-stop next departures |

## Architecture

```
GTFS Static ZIP (daily)          Metrolinx API (10-30s polls)
   assets.metrolinx.com            api.openmetrolinx.com
          │                                │
          ▼                                ▼
   ┌─────────────────────────────────────────┐
   │           Go API ("Six Rail")           │
   │                                         │
   │  StaticStore (in-memory)  RTCache       │
   │  - stops map[id]Stop      - positions   │
   │  - routes map[id]Route    - tripUpdates │
   │  - stopRoutes index       - alerts      │
   │                                         │
   │  Enrichment: RT data + static lookups   │
   │  → typed JSON responses                 │
   └──────────────┬──────────────────────────┘
                  │
                  ▼
          SvelteKit Frontend
```

Static data loads on startup, refreshes daily. Real-time feeds poll in background goroutines. The frontend receives enriched JSON (e.g., vehicle positions include route name and color, not just route_id).

## Features & Endpoints

| Feature | Endpoint | Data Source | Refresh |
|---|---|---|---|
| Stations | `GET /api/stops` | GTFS Static (`stops.txt`) | Daily |
| Live Departures | `GET /api/departures/{stopCode}` | Metrolinx API (REST) | On request, 30s cache |
| Train Map | `GET /api/positions` | Metrolinx API (GTFS-RT VehiclePosition) | 10s background poll |
| Alerts | `GET /api/alerts` | Metrolinx API (GTFS-RT Alerts) | 30s background poll |

Removed: schedule browser, journey planner, fares, union departures, trains at-a-glance, info alerts, exceptions.

## Data Model

```go
// From GTFS Static
type Stop struct {
    ID        string  `json:"id"`
    Code      string  `json:"code"`
    Name      string  `json:"name"`
    Lat       float64 `json:"lat"`
    Lon       float64 `json:"lon"`
    ParentID  string  `json:"parentId,omitempty"`
}

type Route struct {
    ID        string `json:"id"`
    ShortName string `json:"shortName"`
    LongName  string `json:"longName"`
    Color     string `json:"color"`
    TextColor string `json:"textColor"`
    Type      int    `json:"type"` // 2=rail, 3=bus
}

// Enriched from GTFS-RT + static lookups
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

Departures remain passthrough from Metrolinx API (no GTFS-RT equivalent for per-stop next-service).

## Go Library

[`jamespfennell/gtfs`](https://github.com/jamespfennell/gtfs) — handles both static CSV parsing (`ParseStatic`) and GTFS-RT protobuf parsing (`ParseRealtime`) in one dependency.

## File Changes

### New files
- `api/internal/gtfs/static.go` — GTFS ZIP downloader, parser, in-memory store with daily refresh
- `api/internal/gtfs/realtime.go` — GTFS-RT feed poller (positions, alerts), protobuf parsing, enrichment
- `api/internal/models/models.go` — typed structs (Stop, Route, VehiclePosition, Alert)

### Modified files
- `api/cmd/server/main.go` — new route registration, startup GTFS load, background pollers
- `api/internal/handlers/handlers.go` — rewritten handlers returning typed JSON
- `api/internal/config/config.go` — add GTFSStaticURL config
- `web/src/lib/api.ts` — remove dropped endpoints, update types
- `web/src/routes/map/+page.svelte` — use new enriched position format
- `web/src/routes/alerts/+page.svelte` — use new alert format
- `web/src/lib/components/Nav.svelte` — remove schedule/journey links

### Removed files
- `web/src/routes/schedule/` — schedule browser page
- `web/src/routes/journey/` — journey planner page
- Dead handler code for removed endpoints

### Rename
- Go module: `github.com/teclara/gopulse/api` → `github.com/teclara/sixrail/api`
- Frontend title, nav, meta tags → "Six Rail"
- `railway.toml` service names updated

## Config

```
# api/.env
METROLINX_API_KEY=...           # still required for RT + departures
GTFS_STATIC_URL=https://assets.metrolinx.com/raw/upload/Documents/Metrolinx/Open%20Data/GO-GTFS.zip
PORT=8080
ALLOWED_ORIGINS=http://localhost:5173

# web/.env
API_BASE_URL=http://localhost:8080
PUBLIC_MAPBOX_TOKEN=...
```
