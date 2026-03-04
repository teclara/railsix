# Filter Panel Design

## Summary

Floating filter chips on the map to filter vehicle positions by route/line, transport type (train/bus), and status (on time/delayed/cancelled).

## UI

Floating chip rows anchored to bottom-left of the map. A filter icon button toggles visibility. When expanded, three rows:

1. **Transport type** — `Trains` / `Buses` toggle chips with icons
2. **Route/line** — Horizontally scrollable colored chips (route color as background, dimmed when inactive)
3. **Status** — `On Time` / `Delayed` / `Cancelled` chips

Collapsed state shows only the filter icon with a badge count of active filters.

## Data Flow

- Filtering is purely client-side in `updatePositionLayer()`
- Positions array is filtered before creating GeoJSON features
- Route list is derived dynamically from current positions (unique routeNames)

## API Change

Add `routeType` field to `VehiclePosition` model (Go + TypeScript):
- GTFS route_type: 2 = rail, 3 = bus
- Set during enrichment in both real-time and simulated paths

## State & Persistence

- Filter state managed via Svelte `$state` in a new `$lib/stores/filters.ts`
- Persisted to `localStorage` (same pattern as `favorites.ts`)
- Shape: `{ showTrains: boolean, showBuses: boolean, activeRoutes: Set<string>, activeStatuses: Set<string> }`

## Components

- `FilterChips.svelte` — new component rendering the floating chip UI
- Used in `+page.svelte` alongside SearchOverlay and AlertsDropdown

## Files to Modify

- `api/internal/models/models.go` — add `RouteType` to `VehiclePosition`
- `api/internal/gtfs/realtime.go` — set `RouteType` in `EnrichPositions`
- `api/internal/gtfs/simulate.go` — set `RouteType` in `SimulatePositions`
- `web/src/lib/api.ts` — add `routeType` to `VehiclePosition` interface
- `web/src/lib/stores/filters.ts` — new file, localStorage-backed filter store
- `web/src/lib/components/FilterChips.svelte` — new component
- `web/src/routes/+page.svelte` — integrate FilterChips, apply filters in `updatePositionLayer`
