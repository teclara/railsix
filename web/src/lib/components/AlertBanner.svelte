<!-- web/src/lib/components/AlertBanner.svelte -->
<script lang="ts">
	let { alerts = [] }: { alerts: any[] } = $props();
</script>

{#if alerts.length > 0}
	<div class="bg-amber-100 border-b border-amber-300 px-4 py-2 text-sm text-amber-900">
		<strong>Service Alert:</strong> {alerts[0].Message || alerts[0].message || 'Service disruption reported'}
		{#if alerts.length > 1}
			<a href="/alerts" class="underline ml-2">+{alerts.length - 1} more</a>
		{/if}
	</div>
{/if}
