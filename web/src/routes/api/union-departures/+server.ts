import { json } from '@sveltejs/kit';
import { getUnionDepartures } from '$lib/api';

export async function GET() {
	try {
		const departures = await getUnionDepartures();
		return json(departures);
	} catch (err) {
		console.error('[proxy] /api/union-departures failed:', err);
		return json({ error: 'upstream unavailable' }, { status: 502 });
	}
}
