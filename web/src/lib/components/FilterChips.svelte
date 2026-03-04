<script lang="ts">
	import type { VehiclePosition } from '$lib/api';
	import { filters, type FilterState } from '$lib/stores/filters';

	let {
		positions = [],
		filterState
	}: { positions: VehiclePosition[]; filterState: FilterState } = $props();

	let expanded = $state(false);

	let routes = $derived(
		[...new Map(positions.map((p) => [p.routeName, p.routeColor])).entries()]
			.map(([name, color]) => ({ name, color }))
			.sort((a, b) => a.name.localeCompare(b.name))
	);

	let activeFilterCount = $derived(
		(filterState.showTrains ? 0 : 1) +
			(filterState.showBuses ? 0 : 1) +
			filterState.activeRoutes.length +
			filterState.activeStatuses.length
	);

	function isRouteActive(routeName: string): boolean {
		return filterState.activeRoutes.length === 0 || filterState.activeRoutes.includes(routeName);
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

			<button
				class="text-xs text-gray-400 hover:text-gray-600 transition-colors"
				onclick={() => filters.reset()}
			>
				Reset filters
			</button>
		</div>
	{/if}

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
		{#if activeFilterCount > 0}
			<span
				class="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[10px] rounded-full flex items-center justify-center"
			>
				{activeFilterCount}
			</span>
		{/if}
	</button>
</div>
