<!-- web/src/lib/components/DepartureBoard.svelte -->
<script lang="ts">
	import type { Departure } from '$lib/api-client';

	let { departures = [] }: { departures: Departure[] } = $props();
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
					<td class="px-4 py-2 font-medium">
						{#if dep.routeColor}
							<span
								class="inline-block rounded px-1.5 py-0.5 text-white text-xs font-bold"
								style="background-color: #{dep.routeColor}"
							>
								{dep.line || '—'}
							</span>
						{:else}
							{dep.line || '—'}
						{/if}
					</td>
					<td class="px-4 py-2">{dep.destination || '—'}</td>
					<td class="px-4 py-2">{dep.scheduledTime || '—'}</td>
					<td class="px-4 py-2">
						<span
							class={dep.status === 'Cancelled'
								? 'text-red-700 font-medium'
								: dep.status?.startsWith('Delayed')
									? 'text-orange-600 font-medium'
									: 'text-green-600'}
						>
							{dep.status || 'On Time'}
						</span>
					</td>
					<td class="px-4 py-2">{dep.platform || '—'}</td>
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
