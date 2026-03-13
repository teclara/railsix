<script lang="ts">
	import MyCommute from '$lib/components/MyCommute.svelte';
	let { data } = $props();
</script>

<svelte:head>
	{#if data.urlTrip}
		<title
			>{data.urlTrip.fromName} → {data.urlTrip.toName} GO Train — Live Countdown | Rail Six</title
		>
		<meta
			name="description"
			content="Live GO Train countdown from {data.urlTrip.fromName} to {data.urlTrip
				.toName}. Real-time departure times, platform assignments, and delay alerts. Free — no account, no ads, no tracking."
		/>
		<meta
			property="og:title"
			content="{data.urlTrip.fromName} → {data.urlTrip.toName} GO Train — Live Countdown | Rail Six"
		/>
		<meta
			property="og:description"
			content="Live GO Train countdown from {data.urlTrip.fromName} to {data.urlTrip
				.toName}. Real-time departures, platform info, and delay alerts."
		/>
		<meta
			property="og:url"
			content="https://railsix.com/?from={data.urlTrip.fromCode}&to={data.urlTrip.toCode}&dir={data
				.urlTrip.dir}"
		/>
	{:else}
		<title>Real-Time GO Train Departures & Commute Tracker — Rail Six</title>
		<meta
			name="description"
			content="Real-time GO Train departure times, platform info, and delay alerts for the GTA. Save your commute for a live countdown. Free — no account, no ads, no tracking."
		/>
		<meta
			property="og:title"
			content="Real-Time GO Train Departures & Commute Tracker — Rail Six"
		/>
		<meta
			property="og:description"
			content="Real-time GO Train departure times, platform info, and delay alerts for the GTA. Free — no account, no ads, no tracking."
		/>
		<meta property="og:url" content="https://railsix.com/" />
	{/if}
	<meta
		name="keywords"
		content="GO Train, GO Transit, GO train real time, GO train departures, GO train platform, GO train countdown, GO train delay, GO Transit schedule, Union Station GO, GTA commute tracker, Toronto commuter rail, GO train no account, GO train tracker"
	/>
	<meta name="author" content="Teclara Technologies Inc" />
	<meta name="robots" content="index, follow" />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content="Rail Six" />
	<meta property="og:image" content="https://railsix.com/og-image.png" />
	<meta property="og:image:width" content="1200" />
	<meta property="og:image:height" content="630" />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:site" content="@railsix" />
	<meta name="twitter:image" content="https://railsix.com/og-image.png" />
</svelte:head>

<MyCommute stops={data.stops} alerts={data.alerts} />
