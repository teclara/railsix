import { getUnionDepartures } from '$lib/api';

export async function load() {
	try {
		const departures = await getUnionDepartures();
		return { departures: departures ?? [] };
	} catch {
		return { departures: [] };
	}
}
