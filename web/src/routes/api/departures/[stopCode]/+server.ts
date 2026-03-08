import { json } from '@sveltejs/kit';
import { getStopDepartures } from '$lib/api';

const stopCodeRe = /^[A-Za-z0-9]{2,10}$/;

export async function GET({ params, url }) {
	if (!stopCodeRe.test(params.stopCode)) {
		return json({ error: 'invalid stop code' }, { status: 400 });
	}
	try {
		const dest = url.searchParams.get('dest') ?? undefined;
		const departures = await getStopDepartures(params.stopCode, dest);
		return json(departures);
	} catch {
		return json([], { status: 502 });
	}
}
