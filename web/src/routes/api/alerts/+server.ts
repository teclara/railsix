import { json } from '@sveltejs/kit';
import { getAlerts } from '$lib/api';

export async function GET() {
	try {
		const alerts = await getAlerts();
		return json(alerts);
	} catch (err) {
		console.error('[proxy] /api/alerts failed:', err);
		return json({ error: 'upstream unavailable' }, { status: 502 });
	}
}
