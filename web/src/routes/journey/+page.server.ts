// web/src/routes/journey/+page.server.ts
import { getAllStops, getScheduleJourney, getFares } from '$lib/api';

export async function load({ url }) {
	const stops = await getAllStops().catch(() => []);
	const from = url.searchParams.get('from');
	const to = url.searchParams.get('to');
	const date = url.searchParams.get('date');
	const startTime = url.searchParams.get('startTime');

	if (!from || !to || !date || !startTime) {
		return { stops: Array.isArray(stops) ? stops : [], journeys: null, fares: null };
	}

	const [journeys, fares] = await Promise.all([
		getScheduleJourney({ date, from, to, startTime }).catch(() => null),
		getFares(from, to).catch(() => null)
	]);

	return {
		stops: Array.isArray(stops) ? stops : [],
		journeys,
		fares
	};
}
