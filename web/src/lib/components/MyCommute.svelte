<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { commute, notificationPrefs, getActiveDirection } from '$lib/stores/commute';
	import type { CommuteStore } from '$lib/stores/commute';
	import type { Stop } from '$lib/api';
	import type { Alert } from '$lib/api';
	import type { Departure } from '$lib/api-client';
	import { fetchDepartures, fetchAlerts } from '$lib/api-client';
	import { torontoHour } from '$lib/display';
	import { track } from '$lib/track';
	import { untrack } from 'svelte';
	import SplitFlapBoard from './SplitFlapBoard.svelte';
	import CountdownTimer from './CountdownTimer.svelte';
	import AlertBanner from './AlertBanner.svelte';
	import CommuteSetup from './CommuteSetup.svelte';
	import SettingsPanel from './SettingsPanel.svelte';

	let { stops, alerts: initialAlerts }: { stops: Stop[]; alerts: Alert[] } = $props();

	let commuteState = $state<CommuteStore>({ toWork: null, toHome: null });
	const unsubCommute = commute.subscribe((s) => (commuteState = s));

	let directionOverride = $state<'toWork' | 'toHome' | null>(null);
	let activeDirection = $derived(getActiveDirection(directionOverride, commuteState));
	let activeTrip = $derived.by(() =>
		activeDirection === 'toWork' ? commuteState.toWork : commuteState.toHome
	);

	let departures = $state<Departure[]>([]);
	let alerts = $state<Alert[]>(untrack(() => initialAlerts));
	let showSettings = $state(false);

	function greeting(): string {
		const h = torontoHour();
		if (h < 12) return 'Good morning';
		if (h < 17) return 'Good afternoon';
		return 'Good evening';
	}

	function dateStr(): string {
		return new Date().toLocaleDateString('en-CA', {
			weekday: 'long',
			month: 'long',
			day: 'numeric',
			timeZone: 'America/Toronto'
		});
	}

	let nextDeparture = $derived(departures[0] ?? null);

	async function loadDepartures(trip = activeTrip) {
		if (!trip) {
			departures = [];
			return;
		}
		try {
			departures = await fetchDepartures(trip.originCode, trip.destinationCode);
			checkDelayNotification(departures);
		} catch {
			departures = [];
		}
	}

	async function loadAlerts() {
		try {
			alerts = await fetchAlerts();
		} catch {
			// keep existing alerts on error
		}
	}

	let departInterval: ReturnType<typeof setInterval>;
	let alertInterval: ReturnType<typeof setInterval>;

	async function requestNotificationAccess(): Promise<boolean> {
		if (!browser || !('Notification' in window)) {
			notificationPrefs.setEnabled(false);
			return false;
		}
		if (Notification.permission === 'granted') {
			notificationPrefs.setEnabled(true);
			return true;
		}
		if (Notification.permission === 'denied') {
			notificationPrefs.setEnabled(false);
			return false;
		}
		const permission = await Notification.requestPermission();
		const enabled = permission === 'granted';
		notificationPrefs.setEnabled(enabled);
		return enabled;
	}

	function syncNotificationAccess() {
		if (!browser || !('Notification' in window) || Notification.permission !== 'granted') {
			notificationPrefs.setEnabled(false);
		}
	}

	async function showDelayNotification(title: string, body: string) {
		if (!browser || !('Notification' in window) || Notification.permission !== 'granted') return;
		if ('serviceWorker' in navigator) {
			const registration = await navigator.serviceWorker.ready;
			await registration.showNotification(title, {
				body,
				icon: '/icons/icon-192.png',
				badge: '/icons/icon-192.png',
				tag: 'delay-alert'
			});
			return;
		}
		new Notification(title, { body, icon: '/icons/icon-192.png' });
	}

	onMount(() => {
		commute.hydrate();
		notificationPrefs.hydrate();
		syncNotificationAccess();
		mounted = true;
		// Departures load is handled by the $effect reacting to activeTrip after hydrate.
		// Alerts are already loaded via SSR (initialAlerts prop) — skip initial fetch.
		departInterval = setInterval(loadDepartures, 30_000);
		alertInterval = setInterval(loadAlerts, 60_000);
	});

	onDestroy(() => {
		clearInterval(departInterval);
		clearInterval(alertInterval);
		unsubCommute();
	});

	let mounted = $state(false);
	$effect(() => {
		const trip = activeDirection === 'toWork' ? commuteState.toWork : commuteState.toHome;
		if (browser && mounted) {
			lastNotifiedDelay = null;
			void loadDepartures(trip);
		}
	});

	// Pass empty array — AlertBanner shows all alerts when no route filter is provided
	// TODO: store route names in commute trips to enable route-specific filtering
	let activeRouteNames = $derived<string[]>([]);

	async function requestNotifications() {
		const enabled = await requestNotificationAccess();
		if (enabled) {
			track('notifications-enabled');
		}
	}

	let notifPrefs = $state({ enabled: false, thresholdMinutes: 5 });
	let notifEnabled = $derived(notifPrefs.enabled);
	const unsubNotif = notificationPrefs.subscribe((s) => (notifPrefs = s));
	onDestroy(() => unsubNotif());

	let lastNotifiedDelay: number | null = null;

	function checkDelayNotification(deps: Departure[]) {
		if (!notifPrefs.enabled || !deps.length) return;
		const next = deps[0];
		const delay = next.delayMinutes ?? 0;

		if (
			lastNotifiedDelay !== null &&
			delay > lastNotifiedDelay &&
			delay >= notifPrefs.thresholdMinutes
		) {
			void showDelayNotification(
				'Rail Six — Delay Alert',
				`Your ${next.scheduledTime} ${next.line} is now delayed ${delay} min`
			);
		}

		lastNotifiedDelay = delay;
	}
</script>

{#if !commuteState.toWork && !commuteState.toHome}
	<CommuteSetup {stops} />
{:else}
	<div
		class="my-commute bg-[#111] h-[calc(100dvh-60px)] text-white font-mono p-4 flex flex-col justify-center gap-4 max-w-lg mx-auto overflow-hidden"
	>
		<!-- Header -->
		<div class="flex items-start justify-between pt-2">
			<div>
				<h1 class="text-amber-400 font-bold text-base uppercase tracking-widest">Rail Six</h1>
				<p class="text-gray-500 text-xs mt-0.5">{greeting()} &middot; {dateStr()}</p>
			</div>
			<button
				class="text-gray-500 hover:text-white text-lg leading-none p-1"
				onclick={() => (showSettings = true)}
				aria-label="Settings"
			>
				⚙
			</button>
		</div>

		<!-- Direction toggle -->
		<div class="flex rounded overflow-hidden border border-[#2a2a2a]">
			<button
				class="flex-1 py-2 text-xs uppercase tracking-wider transition-colors"
				class:bg-amber-400={activeDirection === 'toWork'}
				class:text-black={activeDirection === 'toWork'}
				class:text-gray-400={activeDirection !== 'toWork'}
				onclick={() => {
					directionOverride = 'toWork';
					track('direction-toggle', { direction: 'toWork' });
				}}
				disabled={!commuteState.toWork}
			>
				To Work
			</button>
			<button
				class="flex-1 py-2 text-xs uppercase tracking-wider transition-colors"
				class:bg-amber-400={activeDirection === 'toHome'}
				class:text-black={activeDirection === 'toHome'}
				class:text-gray-400={activeDirection !== 'toHome'}
				onclick={() => {
					directionOverride = 'toHome';
					track('direction-toggle', { direction: 'toHome' });
				}}
				disabled={!commuteState.toHome}
			>
				To Home
			</button>
		</div>

		{#if activeTrip}
			<!-- Route header -->
			<div class="text-center">
				<p class="text-xs text-gray-500 uppercase tracking-widest">
					{activeTrip.originName} → {activeTrip.destinationName}
				</p>
			</div>
		{/if}

		<!-- Alert banner -->
		<AlertBanner {alerts} routeNames={activeRouteNames} />

		<!-- Split-flap board -->
		<SplitFlapBoard {departures} maxRows={3} />

		<!-- Countdown -->
		{#if nextDeparture}
			<div class="flex justify-center mt-2">
				<CountdownTimer scheduledTime={nextDeparture.scheduledTime} />
			</div>
		{/if}

		<!-- Notification toggle -->
		{#if activeTrip}
			<div class="flex items-center justify-center mt-2">
				{#if notifEnabled}
					<p class="text-green-500 text-xs font-mono">🔔 Delay notifications on</p>
				{:else}
					<button
						class="text-gray-500 text-xs font-mono hover:text-amber-400 transition-colors"
						onclick={requestNotifications}
					>
						🔔 Notify me if delayed
					</button>
				{/if}
			</div>
		{:else}
			<div class="text-center text-gray-600 text-xs py-12">
				No trip configured for this direction.<br />
				<button class="text-amber-400 mt-2" onclick={() => (showSettings = true)}
					>Set up in settings →</button
				>
			</div>
		{/if}

		<footer class="pt-2 pb-4 text-center max-w-xs mx-auto">
			<p class="text-gray-500 text-[11px] font-mono leading-relaxed">
				Real-time GO Transit tracking with live departures, delay alerts, and countdown timers for
				your daily commute.
			</p>
			<p class="text-gray-600 text-[10px] font-mono mt-3 leading-relaxed text-left">
				View live departure times, platform info, and delay notifications for your saved commute.
				Visit the <a href="/departures" class="text-amber-400 hover:text-amber-300 transition-colors"
					>departure board</a
				> for a full split-flap display of upcoming trains at any station.
			</p>
			<p class="text-gray-700 text-[9px] font-mono mt-3 leading-relaxed">
				Not affiliated with Metrolinx or GO Transit. Schedule data may be inaccurate or delayed.
				Always confirm with official sources before travelling.
			</p>
			<p class="text-gray-600 text-[10px] tracking-wide font-mono mt-3">
				&copy; {new Date().getFullYear()}
				<a
					href="https://wadhah.com"
					target="_blank"
					rel="noopener noreferrer"
					class="hover:text-gray-400 transition-colors">Wadhah Hussain</a
				>
			</p>
		</footer>
	</div>

	{#if showSettings}
		<SettingsPanel {stops} onClose={() => (showSettings = false)} />
	{/if}
{/if}
