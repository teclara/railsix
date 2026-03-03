<!-- web/src/routes/+layout.svelte -->
<script>
	import '../app.css';
	import Nav from '$lib/components/Nav.svelte';
	import AlertBanner from '$lib/components/AlertBanner.svelte';

	let { data, children } = $props();
</script>

<div class="min-h-screen bg-gray-50">
	<Nav />
	<AlertBanner alerts={data.alerts} />
	<main class="max-w-6xl mx-auto px-4 py-6">
		{@render children()}
	</main>
</div>
