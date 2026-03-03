import { getStopDepartures } from '$lib/api';

export async function load({ params }) {
	try {
		const departures = await getStopDepartures(params.stopCode);
		return {
			stopCode: params.stopCode,
			departures: Array.isArray(departures) ? departures : []
		};
	} catch {
		return { stopCode: params.stopCode, departures: [] };
	}
}
