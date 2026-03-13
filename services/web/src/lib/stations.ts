import type { Stop } from '$lib/api';

/** Convert a stop name to a URL-friendly slug: "Oakville GO" → "oakville" */
export function stopToSlug(stop: Stop): string {
	return stop.name
		.replace(/\s+GO$/i, '')
		.replace(/\s+Station$/i, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/(^-|-$)/g, '');
}

/** Find a stop by its URL slug. Matches against name-derived slugs. */
export function findStopBySlug(stops: Stop[], slug: string): Stop | undefined {
	return stops.find((s) => stopToSlug(s) === slug);
}
