// web/src/routes/departures/[stopCode]/+page.server.ts
import { getStopDepartures, getStopDetails } from '$lib/api';

export async function load({ params }) {
	const [departures, stopDetails] = await Promise.all([
		getStopDepartures(params.stopCode).catch(() => []),
		getStopDetails(params.stopCode).catch(() => null)
	]);

	return {
		stopCode: params.stopCode,
		departures: Array.isArray(departures) ? departures : [],
		stopDetails: stopDetails as Record<string, unknown> | null
	};
}
