<script lang="ts">
	import MyCommute from '$lib/components/MyCommute.svelte';
	let { data } = $props();
</script>

<svelte:head>
	<title>Rail Six — GO Transit Commute Tracker</title>
	<meta
		name="description"
		content="Real-time GO Transit departure board and commute tracker. Track trains, delays, and platform info for your daily commute."
	/>
	<meta name="author" content="Teclara Technologies Inc" />
	<meta property="og:title" content="Rail Six — GO Transit Commute Tracker" />
	<meta
		property="og:description"
		content="Real-time GO Transit departure board and commute tracker. Track trains, delays, and platform info for your daily commute."
	/>
	<meta property="og:type" content="website" />
	<meta property="og:url" content="https://railsix.com/" />
	<meta property="og:image" content="https://railsix.com/train.png" />
	<meta name="twitter:card" content="summary" />
	<meta name="twitter:title" content="Rail Six — GO Transit Commute Tracker" />
	<meta
		name="twitter:description"
		content="Real-time GO Transit departure board and commute tracker."
	/>
	<meta name="twitter:image" content="https://railsix.com/train.png" />
</svelte:head>

<MyCommute stops={data.stops} alerts={data.alerts} />
