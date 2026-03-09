import type { Departure } from '$lib/api-client';

const torontoFmt = new Intl.DateTimeFormat('en-CA', {
	timeZone: 'America/Toronto',
	hour: 'numeric',
	minute: 'numeric',
	second: 'numeric',
	hour12: false
});

/** Current hour (0-23) in America/Toronto timezone. */
export function torontoHour(): number {
	const parts = torontoFmt.formatToParts(new Date());
	return Number(parts.find((p) => p.type === 'hour')!.value);
}

/**
 * Returns the current time-of-day in America/Toronto as total milliseconds since midnight,
 * plus a helper to build a "today in Toronto" Date for a given HH:MM schedule time.
 */
export function torontoNow(): { ms: number; todayAt: (h: number, m: number) => number } {
	const parts = torontoFmt.formatToParts(new Date());
	const get = (t: string) => Number(parts.find((p) => p.type === t)!.value);
	const hour = get('hour');
	const minute = get('minute');
	const second = get('second');
	const ms = (hour * 3600 + minute * 60 + second) * 1000;
	return {
		ms,
		todayAt: (h: number, m: number) => (h * 3600 + m * 60) * 1000
	};
}

export function padRight(str: string, len: number): string {
	return str.toUpperCase().padEnd(len, ' ').slice(0, len);
}

/** Compact platform strings like "11 & 12" → "11&12" to fit display width. */
export function compactPlatform(p: string): string {
	return p.replace(/\s*&\s*/g, '&');
}

export function padCenter(str: string, len: number): string {
	const s = str.toUpperCase().slice(0, len);
	const left = Math.floor((len - s.length) / 2);
	return s.padStart(s.length + left, ' ').padEnd(len, ' ');
}

export function statusText(d: Departure): string {
	if (d.isCancelled) return 'CANCEL';
	if (d.delayMinutes && d.delayMinutes > 0) return `+${d.delayMinutes}M`;
	return 'ON TIME';
}

export function statusClass(d: Departure): string {
	if (d.isCancelled) return 'text-red-500';
	if (d.delayMinutes && d.delayMinutes > 0) return 'text-amber-400';
	return 'text-green-400';
}
