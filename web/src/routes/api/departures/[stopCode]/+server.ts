import { json } from '@sveltejs/kit';
import { getStopDepartures } from '$lib/api';

export async function GET({ params, url }) {
	try {
		const dest = url.searchParams.get('dest') ?? undefined;
		const departures = await getStopDepartures(params.stopCode, dest);
		return json(departures);
	} catch {
		return json([], { status: 502 });
	}
}
