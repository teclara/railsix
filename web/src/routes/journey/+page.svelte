<!-- web/src/routes/journey/+page.svelte -->
<script lang="ts">
	import { goto } from '$app/navigation';

	let { data } = $props();
	let from = $state('');
	let to = $state('');
	let date = $state(new Date().toISOString().split('T')[0]);
	let startTime = $state('08:00');

	function search() {
		if (!from || !to) return;
		goto(`/journey?from=${from}&to=${to}&date=${date}&startTime=${startTime}`);
	}
</script>

<div class="space-y-6">
	<h1 class="text-2xl font-bold">Journey Planner</h1>

	<div class="bg-white border border-gray-200 rounded-lg p-4 space-y-3">
		<div class="grid grid-cols-2 gap-3">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">From</label>
				<select bind:value={from} class="w-full px-3 py-2 border border-gray-300 rounded">
					<option value="">Select station</option>
					{#each data.stops as stop}
						<option value={(stop as any).StopCode || (stop as any).Code}>{(stop as any).StopName || (stop as any).Name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">To</label>
				<select bind:value={to} class="w-full px-3 py-2 border border-gray-300 rounded">
					<option value="">Select station</option>
					{#each data.stops as stop}
						<option value={(stop as any).StopCode || (stop as any).Code}>{(stop as any).StopName || (stop as any).Name}</option>
					{/each}
				</select>
			</div>
		</div>
		<div class="grid grid-cols-2 gap-3">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Date</label>
				<input type="date" bind:value={date} class="w-full px-3 py-2 border border-gray-300 rounded" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Depart after</label>
				<input type="time" bind:value={startTime} class="w-full px-3 py-2 border border-gray-300 rounded" />
			</div>
		</div>
		<button onclick={search}
			class="w-full bg-green-700 text-white py-2 rounded font-medium hover:bg-green-800">
			Search
		</button>
	</div>

	{#if data.fares}
		<div class="bg-green-50 border border-green-200 rounded-lg p-4">
			<p class="font-medium">Fare Information</p>
			<pre class="text-sm mt-1">{JSON.stringify(data.fares, null, 2)}</pre>
		</div>
	{/if}

	{#if data.journeys}
		<div class="space-y-3">
			<h2 class="text-lg font-medium">Results</h2>
			<pre class="bg-white border rounded-lg p-4 text-sm overflow-x-auto">{JSON.stringify(data.journeys, null, 2)}</pre>
		</div>
	{/if}
</div>
