<!-- web/src/lib/components/DepartureBoard.svelte -->
<script lang="ts">
	let { departures = [] }: { departures: any[] } = $props();
</script>

<div class="overflow-x-auto">
	<table class="w-full text-sm">
		<thead class="bg-green-700 text-white">
			<tr>
				<th class="px-4 py-2 text-left">Line</th>
				<th class="px-4 py-2 text-left">Destination</th>
				<th class="px-4 py-2 text-left">Scheduled</th>
				<th class="px-4 py-2 text-left">Status</th>
				<th class="px-4 py-2 text-left">Platform</th>
			</tr>
		</thead>
		<tbody>
			{#each departures as dep}
				<tr class="border-b border-gray-200 hover:bg-gray-50">
					<td class="px-4 py-2 font-medium">{dep.LineName || dep.Line || '—'}</td>
					<td class="px-4 py-2">{dep.Destination || dep.DirectionName || '—'}</td>
					<td class="px-4 py-2">{dep.ScheduledTime || dep.Time || '—'}</td>
					<td class="px-4 py-2">
						<span class={dep.Late || dep.Delayed ? 'text-red-600 font-medium' : 'text-green-600'}>
							{dep.Status || (dep.Late ? 'Delayed' : 'On Time')}
						</span>
					</td>
					<td class="px-4 py-2">{dep.Platform || dep.Track || '—'}</td>
				</tr>
			{:else}
				<tr>
					<td colspan="5" class="px-4 py-8 text-center text-gray-500">
						No active departures at this time.
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
