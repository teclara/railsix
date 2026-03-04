# Filter Panel Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add floating filter chips to the map that let users filter vehicle positions by route/line, transport type (train/bus), and status (on time/delayed/cancelled).

**Architecture:** Client-side filtering applied in `updatePositionLayer()` before generating GeoJSON. Filter state persisted to localStorage. API extended to include `routeType` in position responses so the client can distinguish trains from buses.

**Tech Stack:** Go (API), SvelteKit 2 / Svelte 5 runes, Tailwind CSS 4

---

### Task 1: Add routeType to VehiclePosition model (Go)

**Files:**
- Modify: `api/internal/models/models.go:21-32`

**Step 1: Add RouteType field**

In `models.go`, add `RouteType` to `VehiclePosition`:

```go
type VehiclePosition struct {
	VehicleID  string  `json:"vehicleId"`
	TripID     string  `json:"tripId"`
	RouteID    string  `json:"routeId"`
	RouteName  string  `json:"routeName"`
	RouteColor string  `json:"routeColor"`
	RouteType  int     `json:"routeType"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Bearing    float32 `json:"bearing,omitempty"`
	Speed      float32 `json:"speed,omitempty"`
	Timestamp  int64   `json:"timestamp"`
}
```

**Step 2: Commit**

```bash
git add api/internal/models/models.go
git commit -m "feat(models): add RouteType to VehiclePosition"
```

---

### Task 2: Set routeType in enrichment and simulation

**Files:**
- Modify: `api/internal/models/models.go:12-19` (Route struct — already has Type)
- Modify: `api/internal/gtfs/realtime.go:214-234` (EnrichPositions)
- Modify: `api/internal/gtfs/simulate.go:34-46` (SimulatePositions loop)

**Step 1: Set RouteType in EnrichPositions**

In `realtime.go` `EnrichPositions`, after setting `RouteName` and `RouteColor`, also set `RouteType`:

```go
if route, ok := lookup.GetRoute(rp.RouteID); ok {
	vp.RouteName = route.LongName
	vp.RouteColor = route.Color
	vp.RouteType = route.Type
}
```

**Step 2: Set RouteType in SimulatePositions**

In `simulate.go`, inside the loop where positions are built, after getting the route:

```go
route, _ := static.GetRoute(trip.RouteID)
positions = append(positions, models.VehiclePosition{
	VehicleID:  trip.TripID,
	TripID:     trip.TripID,
	RouteID:    trip.RouteID,
	RouteName:  route.LongName,
	RouteColor: route.Color,
	RouteType:  route.Type,
	Lat:        pos.lat,
	Lon:        pos.lon,
	Bearing:    pos.bearing,
	Timestamp:  now.Unix(),
})
```

**Step 3: Commit**

```bash
git add api/internal/gtfs/realtime.go api/internal/gtfs/simulate.go
git commit -m "feat(gtfs): set routeType in position enrichment and simulation"
```

---

### Task 3: Add routeType to TypeScript interface

**Files:**
- Modify: `web/src/lib/api.ts:29-40`

**Step 1: Add routeType to VehiclePosition interface**

```typescript
export interface VehiclePosition {
	vehicleId: string;
	tripId: string;
	routeId: string;
	routeName: string;
	routeColor: string;
	routeType: number;
	lat: number;
	lon: number;
	bearing?: number;
	speed?: number;
	timestamp: number;
}
```

**Step 2: Commit**

```bash
git add web/src/lib/api.ts
git commit -m "feat(web): add routeType to VehiclePosition interface"
```

---

### Task 4: Create filter store with localStorage persistence

**Files:**
- Create: `web/src/lib/stores/filters.ts`

**Step 1: Create the filter store**

Follow the same pattern as `favorites.ts`. The store holds transport toggles, selected routes, and selected statuses.

```typescript
import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export interface FilterState {
	showTrains: boolean;
	showBuses: boolean;
	activeRoutes: string[]; // empty = show all
	activeStatuses: string[]; // 'ontime' | 'delayed' | 'cancelled'; empty = show all
}

const defaultFilters: FilterState = {
	showTrains: true,
	showBuses: true,
	activeRoutes: [],
	activeStatuses: []
};

function createFilters() {
	const initial: FilterState = browser
		? { ...defaultFilters, ...JSON.parse(localStorage.getItem('filters') || '{}') }
		: defaultFilters;
	const { subscribe, set, update } = writable<FilterState>(initial);

	function persist(state: FilterState) {
		if (browser) localStorage.setItem('filters', JSON.stringify(state));
	}

	return {
		subscribe,
		toggleTrains() {
			update((s) => {
				const next = { ...s, showTrains: !s.showTrains };
				persist(next);
				return next;
			});
		},
		toggleBuses() {
			update((s) => {
				const next = { ...s, showBuses: !s.showBuses };
				persist(next);
				return next;
			});
		},
		toggleRoute(routeName: string) {
			update((s) => {
				const routes = s.activeRoutes.includes(routeName)
					? s.activeRoutes.filter((r) => r !== routeName)
					: [...s.activeRoutes, routeName];
				const next = { ...s, activeRoutes: routes };
				persist(next);
				return next;
			});
		},
		setAllRoutes(routeNames: string[]) {
			update((s) => {
				const next = { ...s, activeRoutes: routeNames };
				persist(next);
				return next;
			});
		},
		clearRoutes() {
			update((s) => {
				const next = { ...s, activeRoutes: [] };
				persist(next);
				return next;
			});
		},
		toggleStatus(status: string) {
			update((s) => {
				const statuses = s.activeStatuses.includes(status)
					? s.activeStatuses.filter((st) => st !== status)
					: [...s.activeStatuses, status];
				const next = { ...s, activeStatuses: statuses };
				persist(next);
				return next;
			});
		},
		reset() {
			persist(defaultFilters);
			set(defaultFilters);
		}
	};
}

export const filters = createFilters();
```

**Step 2: Commit**

```bash
git add web/src/lib/stores/filters.ts
git commit -m "feat(web): add localStorage-backed filter store"
```

---

### Task 5: Create FilterChips component

**Files:**
- Create: `web/src/lib/components/FilterChips.svelte`

**Step 1: Create the component**

Floating chip UI with three rows: transport toggles, route chips (scrollable), status chips. Uses Svelte 5 runes (`$props`, `$derived`). Tailwind CSS for styling.

```svelte
<script lang="ts">
	import type { VehiclePosition } from '$lib/api';
	import { filters, type FilterState } from '$lib/stores/filters';

	let {
		positions = [],
		filterState
	}: { positions: VehiclePosition[]; filterState: FilterState } = $props();

	let expanded = $state(false);

	// Derive unique routes from positions
	let routes = $derived(
		[...new Map(positions.map((p) => [p.routeName, p.routeColor])).entries()]
			.map(([name, color]) => ({ name, color }))
			.sort((a, b) => a.name.localeCompare(b.name))
	);

	// Count active filters for badge
	let activeFilterCount = $derived(() => {
		let count = 0;
		if (!filterState.showTrains) count++;
		if (!filterState.showBuses) count++;
		if (filterState.activeRoutes.length > 0) count += filterState.activeRoutes.length;
		if (filterState.activeStatuses.length > 0) count += filterState.activeStatuses.length;
		return count;
	});

	function isRouteActive(routeName: string): boolean {
		return (
			filterState.activeRoutes.length === 0 || filterState.activeRoutes.includes(routeName)
		);
	}

	const statuses = [
		{ key: 'ontime', label: 'On Time' },
		{ key: 'delayed', label: 'Delayed' },
		{ key: 'cancelled', label: 'Cancelled' }
	];

	function isStatusActive(key: string): boolean {
		return filterState.activeStatuses.length === 0 || filterState.activeStatuses.includes(key);
	}
</script>

<div class="absolute bottom-6 left-3 z-20 flex flex-col items-start gap-2">
	{#if expanded}
		<div
			class="bg-white/95 backdrop-blur-sm rounded-xl shadow-lg p-3 max-w-[calc(100vw-1.5rem)] space-y-2"
		>
			<!-- Transport type -->
			<div class="flex gap-1.5">
				<button
					class="px-3 py-1.5 rounded-full text-xs font-medium transition-all {filterState.showTrains
						? 'bg-green-700 text-white'
						: 'bg-gray-200 text-gray-500'}"
					onclick={() => filters.toggleTrains()}
				>
					Train
				</button>
				<button
					class="px-3 py-1.5 rounded-full text-xs font-medium transition-all {filterState.showBuses
						? 'bg-blue-600 text-white'
						: 'bg-gray-200 text-gray-500'}"
					onclick={() => filters.toggleBuses()}
				>
					Bus
				</button>
			</div>

			<!-- Route chips -->
			<div class="flex gap-1.5 overflow-x-auto max-w-[80vw] pb-1">
				{#each routes as route}
					<button
						class="px-2.5 py-1 rounded-full text-xs font-medium whitespace-nowrap transition-all border"
						style={isRouteActive(route.name)
							? `background-color: #${route.color || '15803d'}; color: white; border-color: transparent;`
							: `background-color: #f3f4f6; color: #9ca3af; border-color: #e5e7eb;`}
						onclick={() => filters.toggleRoute(route.name)}
					>
						{route.name}
					</button>
				{/each}
			</div>

			<!-- Status chips -->
			<div class="flex gap-1.5">
				{#each statuses as status}
					<button
						class="px-3 py-1.5 rounded-full text-xs font-medium transition-all {isStatusActive(status.key)
							? 'bg-gray-800 text-white'
							: 'bg-gray-200 text-gray-500'}"
						onclick={() => filters.toggleStatus(status.key)}
					>
						{status.label}
					</button>
				{/each}
			</div>

			<!-- Reset -->
			<button
				class="text-xs text-gray-400 hover:text-gray-600 transition-colors"
				onclick={() => filters.reset()}
			>
				Reset filters
			</button>
		</div>
	{/if}

	<!-- Toggle button -->
	<button
		class="w-10 h-10 rounded-full bg-white shadow-lg flex items-center justify-center hover:bg-gray-50 transition-colors relative"
		onclick={() => (expanded = !expanded)}
		aria-label="Toggle filters"
	>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="w-5 h-5 text-gray-700"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			stroke-width="2"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
			/>
		</svg>
		{#if activeFilterCount() > 0}
			<span
				class="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[10px] rounded-full flex items-center justify-center"
			>
				{activeFilterCount()}
			</span>
		{/if}
	</button>
</div>
```

**Step 2: Commit**

```bash
git add web/src/lib/components/FilterChips.svelte
git commit -m "feat(web): add FilterChips component"
```

---

### Task 6: Integrate filters into the map page

**Files:**
- Modify: `web/src/routes/+page.svelte`

**Step 1: Import FilterChips and filter store**

Add imports at the top of the `<script>` block:

```typescript
import FilterChips from '$lib/components/FilterChips.svelte';
import { filters } from '$lib/stores/filters';
```

Add filter state variable after the existing `$state` declarations:

```typescript
let filterState = $state($filters);
```

Subscribe to filter store changes:

```typescript
// Inside the script block, after state declarations
$effect(() => {
	const unsub = filters.subscribe((s) => (filterState = s));
	return unsub;
});
```

**Step 2: Apply filters in updatePositionLayer**

Replace the `updatePositionLayer` function to filter positions before creating GeoJSON:

```typescript
function updatePositionLayer() {
	if (!map || !mapReady) return;
	const source = map.getSource('positions');
	if (!source) return;

	const filtered = positions.filter((p) => {
		if (!p.lat || !p.lon) return false;

		// Transport type filter
		const isRail = p.routeType === 2 || p.routeType === 0; // 0 = unknown, treat as rail
		const isBus = p.routeType === 3;
		if (isRail && !filterState.showTrains) return false;
		if (isBus && !filterState.showBuses) return false;

		// Route filter (empty = show all)
		if (
			filterState.activeRoutes.length > 0 &&
			!filterState.activeRoutes.includes(p.routeName)
		)
			return false;

		// Status filter (empty = show all) — for now all simulated are 'ontime'
		if (filterState.activeStatuses.length > 0) {
			const status = 'ontime'; // TODO: derive from real-time delay data when available
			if (!filterState.activeStatuses.includes(status)) return false;
		}

		return true;
	});

	source.setData({
		type: 'FeatureCollection',
		features: filtered.map((p) => ({
			type: 'Feature',
			geometry: { type: 'Point', coordinates: [p.lon, p.lat] },
			properties: {
				routeName: p.routeName || p.routeId || '',
				tripId: p.tripId || '',
				color: p.routeColor ? `#${p.routeColor}` : '#15803d'
			}
		}))
	});
}
```

**Step 3: Add the $effect to rerun updatePositionLayer when filterState changes**

Update the existing `$effect` that watches positions to also watch filterState:

```typescript
$effect(() => {
	positions;
	filterState;
	updatePositionLayer();
});
```

**Step 4: Add FilterChips to the template**

In the template, after the `<AlertsDropdown>` line, add:

```svelte
<FilterChips {positions} {filterState} />
```

**Step 5: Commit**

```bash
git add web/src/routes/+page.svelte
git commit -m "feat(web): integrate filter chips into map page"
```

---

### Task 7: Deploy and verify

**Step 1: Push to remote**

```bash
git push
```

**Step 2: Redeploy API**

```bash
railway redeploy -s sixrail-api --yes
```

Wait for SUCCESS, then check logs for `routeType` in position output.

**Step 3: Redeploy web**

```bash
railway redeploy -s sixrail-web --yes
```

**Step 4: Verify**

- Open https://sixrail.up.railway.app
- Click the filter icon (bottom-left)
- Toggle Trains off — dots should disappear
- Toggle Trains back on — dots return
- Click a route chip — only that route shows
- Click Reset — all filters cleared
