<script lang="ts">
	import { onMount, onDestroy, untrack } from 'svelte';
	import { fetchUnionDepartures, type UnionDeparture } from '$lib/api-client';
	import SplitFlapChar from '$lib/components/SplitFlapChar.svelte';

	let { data }: { data: { departures: UnionDeparture[] } } = $props();

	let departures = $state<UnionDeparture[]>(untrack(() => [...data.departures]));
	let clock = $state('');
	let clockInterval: ReturnType<typeof setInterval>;
	let pollInterval: ReturnType<typeof setInterval>;

	function updateClock() {
		const now = new Date();
		clock = now.toLocaleTimeString('en-CA', {
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
			hour12: false,
			timeZone: 'America/Toronto'
		});
	}

	async function loadDepartures() {
		const data = await fetchUnionDepartures();
		if (data.length > 0) departures = data;
	}

	onMount(() => {
		updateClock();
		clockInterval = setInterval(updateClock, 1000);
		pollInterval = setInterval(loadDepartures, 30_000);
	});

	onDestroy(() => {
		clearInterval(clockInterval);
		clearInterval(pollInterval);
	});

	function padRight(str: string, len: number): string {
		return str.toUpperCase().padEnd(len, ' ').slice(0, len);
	}

	function infoClass(info: string): string {
		if (info.includes('PROCEED')) return 'text-green-400';
		if (info.includes('WAIT')) return 'text-amber-400';
		if (info.includes('CANCEL')) return 'text-red-500';
		return 'text-gray-400';
	}
</script>

<svelte:head>
	<title>Union Station — Six Rail</title>
</svelte:head>

<div class="board font-mono select-none bg-[#0a0a0a] min-h-screen text-white">
	<!-- Header -->
	<div class="board-header px-4 py-3 flex items-center justify-between border-b border-[#1a1a1a]">
		<div>
			<h1 class="text-amber-400 text-sm font-bold uppercase tracking-[0.2em]">Union Station GO</h1>
			<p class="text-gray-600 text-xs tracking-widest uppercase">Departures</p>
		</div>
		<div class="text-amber-400 text-lg tracking-widest tabular-nums">{clock}</div>
	</div>

	<!-- Column headers -->
	<div class="col-headers px-4 pt-3 pb-2 text-gray-600 text-xs uppercase tracking-widest border-b border-[#161616]">
		<span class="col-time">Time</span>
		<span class="col-service">Service</span>
		<span class="col-plat">Plat</span>
		<span class="col-info">Status</span>
	</div>

	<!-- Departure rows -->
	<div class="rows px-4">
		{#each departures as dep, i}
			<div class="departure-row" class:first={i === 0}>
				<!-- Split-flap chars row -->
				<div class="flap-row">
					<span class="col-time text-amber-400">
						{#each padRight(dep.time, 5).split('') as char, j}
							<SplitFlapChar value={char} delay={j * 25} />
						{/each}
					</span>

					<span class="col-service text-white">
						{#each padRight(dep.service, 16).split('') as char, j}
							<SplitFlapChar value={char} delay={30 + j * 18} />
						{/each}
					</span>

					<span class="col-plat text-white">
						{#each padRight(dep.platform || '--', 5).split('') as char, j}
							<SplitFlapChar value={char} delay={80 + j * 20} />
						{/each}
					</span>

					<span class="col-info {infoClass(dep.info)}">
						{#each padRight(dep.info, 7).split('') as char, j}
							<SplitFlapChar value={char} delay={100 + j * 18} />
						{/each}
					</span>
				</div>

				<!-- Stops sub-line -->
				{#if dep.stops.length > 0}
					<div class="stops-line text-gray-600 text-xs tracking-wide truncate pl-[calc(5ch+2px+8px)]">
						{dep.stops.join(' · ')}
					</div>
				{/if}
			</div>
		{/each}

		{#if departures.length === 0}
			<div class="text-gray-700 font-mono text-sm py-16 text-center tracking-widest uppercase">
				No departures
			</div>
		{/if}
	</div>
</div>

<style>
	.col-headers,
	.flap-row {
		display: grid;
		grid-template-columns: 5ch 16ch 5ch 7ch;
		gap: 8px;
		align-items: center;
	}

	.col-time,
	.col-service,
	.col-plat,
	.col-info {
		display: flex;
		flex-wrap: nowrap;
		align-items: center;
		overflow: hidden;
	}

	.col-service {
		font-size: 0.85em;
	}
	.col-plat {
		font-size: 0.85em;
	}
	.col-info {
		font-size: 0.8em;
	}

	.departure-row {
		border-bottom: 1px solid #161616;
		padding: 8px 0;
	}

	.departure-row.first {
		border-bottom-color: #222;
		padding: 10px 0;
	}

	.stops-line {
		margin-top: 3px;
	}

	@media (max-width: 480px) {
		.col-headers,
		.flap-row {
			grid-template-columns: 5ch 13ch 4ch 7ch;
			gap: 4px;
		}
	}
</style>
