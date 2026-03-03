<!-- web/src/routes/schedule/+page.svelte -->
<script lang="ts">
	import { goto } from '$app/navigation';

	let { data } = $props();
	let date = $state(data.date);

	function changeDate() {
		goto(`/schedule?date=${date}`);
	}
</script>

<div class="space-y-4">
	<h1 class="text-2xl font-bold">Schedule</h1>

	<div class="flex gap-2 items-center">
		<input type="date" bind:value={date} onchange={changeDate}
			class="px-3 py-2 border border-gray-300 rounded" />
	</div>

	<div class="grid gap-3">
		{#each data.lines as line}
			<div class="bg-white border border-gray-200 rounded-lg p-4">
				<p class="font-medium">{(line as any).LineName || (line as any).Name || 'Line'}</p>
				<p class="text-sm text-gray-500">{(line as any).Direction || ''}</p>
			</div>
		{:else}
			<p class="text-gray-500 py-8 text-center">No lines found for this date.</p>
		{/each}
	</div>
</div>
