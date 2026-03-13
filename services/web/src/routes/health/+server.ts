import { json } from '@sveltejs/kit';

import { getPublicHealth } from '$lib/server/health';

export async function GET() {
	const { status, body } = await getPublicHealth();
	return json(body, { status });
}
