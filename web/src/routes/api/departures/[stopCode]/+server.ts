import { json } from '@sveltejs/kit';
import { getStopDepartures } from '$lib/api';

const stopCodeRe = /^[A-Za-z0-9]{2,10}$/;

export async function GET({ params, url }) {
	if (!stopCodeRe.test(params.stopCode)) {
		return json({ error: 'invalid stop code' }, { status: 400 });
	}
	const dest = url.searchParams.get('dest') ?? undefined;
	if (dest !== undefined && !stopCodeRe.test(dest)) {
		return json({ error: 'invalid dest code' }, { status: 400 });
	}
	try {
		const departures = await getStopDepartures(params.stopCode, dest);
		return json(departures);
	} catch (err) {
		console.error('[proxy] /api/departures failed:', err);
		return json({ error: 'upstream unavailable' }, { status: 502 });
	}
}
