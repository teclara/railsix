import { json } from '@sveltejs/kit';
import { getAllStops } from '$lib/api';

export async function GET() {
	try {
		const stops = await getAllStops();
		return json(stops);
	} catch (err) {
		console.error('[proxy] /api/stops failed:', err);
		return json({ error: 'upstream unavailable' }, { status: 502 });
	}
}
