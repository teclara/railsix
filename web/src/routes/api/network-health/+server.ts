import { json } from '@sveltejs/kit';
import { getNetworkHealth } from '$lib/api';

export async function GET() {
	try {
		const lines = await getNetworkHealth();
		return json(lines);
	} catch (err) {
		console.error('[proxy] /api/network-health failed:', err);
		return json({ error: 'upstream unavailable' }, { status: 502 });
	}
}
