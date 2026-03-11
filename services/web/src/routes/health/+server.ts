import { json } from '@sveltejs/kit';

import { getWebHealth } from '$lib/server/health';

export async function GET() {
	const { status, body } = await getWebHealth();
	return json(body, { status });
}
