<!-- web/src/lib/components/Nav.svelte -->
<script lang="ts">
	import { page } from '$app/stores';

	let pathname = $derived($page.url.pathname as string);
</script>

<nav class="bg-green-700 text-white">
	<div class="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
		<a href="/" class="text-xl font-bold tracking-tight">GoPulse</a>
		<div class="flex gap-4 text-sm">
			<a href="/stations" class:font-bold={pathname === '/stations'}>Stations</a>
			<a href="/map" class:font-bold={pathname === '/map'}>Map</a>
			<a href="/schedule" class:font-bold={pathname === '/schedule'}>Schedule</a>
			<a href="/journey" class:font-bold={pathname === '/journey'}>Journey</a>
			<a href="/alerts" class:font-bold={pathname === '/alerts'}>Alerts</a>
		</div>
	</div>
</nav>
