<script lang="ts">
	import type { VehiclePosition } from '$lib/api';
	import { filters, type FilterState } from '$lib/stores/filters';

	let { positions = [], filterState }: { positions: VehiclePosition[]; filterState: FilterState } =
		$props();

	let expanded = $state(false);

	let trainCount = $derived(positions.filter((p) => p.routeType === 2 || p.routeType === 0).length);
	let busCount = $derived(positions.filter((p) => p.routeType === 3).length);

	let routes = $derived(
		[...new Map(positions.map((p) => [p.routeName, p.routeColor])).entries()]
			.map(([name, color]) => ({ name, color }))
			.sort((a, b) => a.name.localeCompare(b.name))
	);

	function isRouteActive(routeName: string): boolean {
		return filterState.activeRoutes.length === 0 || filterState.activeRoutes.includes(routeName);
	}
</script>

<div class="absolute bottom-6 left-3 z-20 flex flex-col items-start gap-2">
	{#if expanded}
		<div
			class="bg-white/95 backdrop-blur-sm rounded-xl shadow-lg p-3 max-w-[calc(100vw-1.5rem)] space-y-2"
		>
			<p class="text-[10px] uppercase tracking-wider text-gray-400 font-semibold">Routes</p>
			<div class="flex gap-1.5 flex-wrap max-w-[80vw] max-h-40 overflow-y-auto pb-1">
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

			<div class="flex items-center justify-between pt-1">
				<button
					class="text-xs text-gray-400 hover:text-gray-600 transition-colors"
					onclick={() => filters.reset()}
				>
					Reset filters
				</button>
				<button
					class="text-xs text-gray-400 hover:text-gray-600 transition-colors"
					onclick={() => (expanded = false)}
				>
					Close
				</button>
			</div>
		</div>
	{/if}

	<!-- Always-visible vehicle type selector -->
	<div class="flex items-center gap-1.5">
		<button
			class="h-10 px-4 rounded-full shadow-lg flex items-center gap-2 transition-all text-sm font-medium {filterState.showTrains
				? 'bg-green-700 text-white'
				: 'bg-white text-gray-400 hover:bg-gray-50'}"
			onclick={() => filters.toggleTrains()}
		>
			<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M4 15V5a3 3 0 013-3h10a3 3 0 013 3v10M4 15l-2 4h20l-2-4M4 15h16" />
				<circle cx="8.5" cy="18.5" r="1.5" fill="currentColor" />
				<circle cx="15.5" cy="18.5" r="1.5" fill="currentColor" />
				<path d="M9 6h6M9 10h6" />
			</svg>
			Trains
			<span
				class="text-xs rounded-full px-1.5 py-0.5 {filterState.showTrains
					? 'bg-green-800/40'
					: 'bg-gray-200'}"
			>
				{trainCount}
			</span>
		</button>

		<button
			class="h-10 px-4 rounded-full shadow-lg flex items-center gap-2 transition-all text-sm font-medium {filterState.showBuses
				? 'bg-blue-600 text-white'
				: 'bg-white text-gray-400 hover:bg-gray-50'}"
			onclick={() => filters.toggleBuses()}
		>
			<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M3 17V7a4 4 0 014-4h10a4 4 0 014 4v10M3 17l-1 2h20l-1-2M3 17h18" />
				<circle cx="7" cy="19.5" r="1.5" fill="currentColor" />
				<circle cx="17" cy="19.5" r="1.5" fill="currentColor" />
				<path d="M5 7h14v5H5z" />
			</svg>
			Buses
			<span
				class="text-xs rounded-full px-1.5 py-0.5 {filterState.showBuses
					? 'bg-blue-700/40'
					: 'bg-gray-200'}"
			>
				{busCount}
			</span>
		</button>

		<button
			class="w-10 h-10 rounded-full bg-white shadow-lg flex items-center justify-center hover:bg-gray-50 transition-colors relative"
			onclick={() => (expanded = !expanded)}
			aria-label="Filter by route"
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
			{#if filterState.activeRoutes.length > 0}
				<span
					class="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[10px] rounded-full flex items-center justify-center"
				>
					{filterState.activeRoutes.length}
				</span>
			{/if}
		</button>
	</div>
</div>
