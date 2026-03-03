<!-- web/src/routes/alerts/+page.svelte -->
<script lang="ts">
	let { data } = $props();
	let filter = $state('all');
</script>

<div class="space-y-4">
	<h1 class="text-2xl font-bold">Service Alerts</h1>

	<div class="flex gap-2">
		<button onclick={() => filter = 'all'} class="px-3 py-1 rounded {filter === 'all' ? 'bg-green-700 text-white' : 'bg-gray-200'}">All</button>
		<button onclick={() => filter = 'service'} class="px-3 py-1 rounded {filter === 'service' ? 'bg-red-600 text-white' : 'bg-gray-200'}">Service</button>
		<button onclick={() => filter = 'info'} class="px-3 py-1 rounded {filter === 'info' ? 'bg-blue-600 text-white' : 'bg-gray-200'}">Info</button>
		<button onclick={() => filter = 'exceptions'} class="px-3 py-1 rounded {filter === 'exceptions' ? 'bg-amber-600 text-white' : 'bg-gray-200'}">Cancellations</button>
	</div>

	{#if filter === 'all' || filter === 'service'}
		{#each data.serviceAlerts as alert}
			<div class="bg-red-50 border-l-4 border-red-500 p-4 rounded">
				<p class="font-medium text-red-900">{(alert as any).Message || (alert as any).Title || 'Service disruption'}</p>
				{#if (alert as any).UpdatedTime}<p class="text-sm text-red-700 mt-1">{(alert as any).UpdatedTime}</p>{/if}
			</div>
		{/each}
	{/if}

	{#if filter === 'all' || filter === 'info'}
		{#each data.infoAlerts as alert}
			<div class="bg-blue-50 border-l-4 border-blue-500 p-4 rounded">
				<p class="font-medium text-blue-900">{(alert as any).Message || (alert as any).Title || 'Information'}</p>
				{#if (alert as any).UpdatedTime}<p class="text-sm text-blue-700 mt-1">{(alert as any).UpdatedTime}</p>{/if}
			</div>
		{/each}
	{/if}

	{#if filter === 'all' || filter === 'exceptions'}
		{#each data.exceptions as exc}
			<div class="bg-amber-50 border-l-4 border-amber-500 p-4 rounded">
				<p class="font-medium text-amber-900">{(exc as any).Message || (exc as any).Title || 'Schedule exception'}</p>
				{#if (exc as any).TripNumber}<p class="text-sm text-amber-700 mt-1">Trip: {(exc as any).TripNumber}</p>{/if}
			</div>
		{/each}
	{/if}

	{#if data.serviceAlerts.length === 0 && data.infoAlerts.length === 0 && data.exceptions.length === 0}
		<p class="text-gray-500 py-8 text-center">No active alerts.</p>
	{/if}
</div>
