import { browser } from '$app/environment';

export type AnalyticsEvent =
	| 'landing_viewed'
	| 'station_selected'
	| 'countdown_viewed'
	| 'saved_commute'
	| 'commute_loaded'
	| 'commute_cleared'
	| 'direction-toggle'
	| 'add_to_home_screen_prompt_seen'
	| 'add_to_home_screen_completed'
	| 'empty_state_viewed'
	| 'error_viewed';

export type TrackProps = Record<string, string | number | boolean>;

function getReferrerGroup(): string {
	if (!browser) return 'unknown';
	const ref = document.referrer;
	if (!ref) return 'direct';
	if (/google\./i.test(ref)) return 'google';
	if (/reddit\.com/i.test(ref)) return 'reddit';
	if (/linkedin\.com/i.test(ref)) return 'linkedin';
	if (/t\.co|twitter\.com|x\.com/i.test(ref)) return 'x';
	if (/facebook\.com|instagram\.com|tiktok\.com/i.test(ref)) return 'social';
	if (/mail\.|webmail\.|outlook\.|gmail\./i.test(ref)) return 'email';
	if (/hacker|lobste|tildes|lemmy/i.test(ref)) return 'community';
	return 'unknown';
}

function getUtmParams(): TrackProps {
	if (!browser) return {};
	const p = new URL(window.location.href).searchParams;
	const result: TrackProps = {};
	const src = p.get('utm_source');
	const med = p.get('utm_medium');
	const cam = p.get('utm_campaign');
	if (src) result['utm_source'] = src;
	if (med) result['utm_medium'] = med;
	if (cam) result['utm_campaign'] = cam;
	return result;
}

function getDeviceType(): 'mobile' | 'tablet' | 'desktop' {
	if (!browser) return 'desktop';
	const w = window.innerWidth;
	if (w < 480) return 'mobile';
	if (w < 1024) return 'tablet';
	return 'desktop';
}

function getViewportBucket(): 'small' | 'medium' | 'large' {
	if (!browser) return 'large';
	const w = window.innerWidth;
	if (w < 480) return 'small';
	if (w < 1024) return 'medium';
	return 'large';
}

export function getLeadTimeBucket(minutes: number): '0_5' | '6_10' | '11_20' | '21_plus' {
	if (minutes <= 5) return '0_5';
	if (minutes <= 10) return '6_10';
	if (minutes <= 20) return '11_20';
	return '21_plus';
}

export function track(event: AnalyticsEvent, props?: TrackProps): void {
	if (!browser) return;
	const enriched: TrackProps = {
		device_type: getDeviceType(),
		viewport: getViewportBucket(),
		referrer_group: getReferrerGroup(),
		...getUtmParams(),
		...props
	};
	window.umami?.track(event, enriched);
}

let countdownViewedThisSession = false;

export function trackCountdownViewed(props: TrackProps): void {
	if (countdownViewedThisSession) return;
	countdownViewedThisSession = true;
	track('countdown_viewed', props);
}
