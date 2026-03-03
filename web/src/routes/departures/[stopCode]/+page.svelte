<!-- web/src/routes/departures/[stopCode]/+page.svelte -->
<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { onMount } from 'svelte';
	import DepartureBoard from '$lib/components/DepartureBoard.svelte';
	import { favorites } from '$lib/stores/favorites';

	let { data } = $props();
	let isFavorite = $derived($favorites.includes(data.stopCode));

	onMount(() => {
		const interval = setInterval(() => invalidateAll(), 30_000);
		return () => clearInterval(interval);
	});
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">
			{data.stopDetails?.StopName || data.stopDetails?.Name || `Station ${data.stopCode}`}
		</h1>
		<button
			onclick={() => favorites.toggle(data.stopCode)}
			class="text-2xl"
			aria-label={isFavorite ? 'Remove from favorites' : 'Add to favorites'}
		>
			{isFavorite ? '★' : '☆'}
		</button>
	</div>
	<p class="text-sm text-gray-500">Auto-refreshes every 30 seconds</p>
	<DepartureBoard departures={data.departures} />
</div>
