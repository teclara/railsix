import { error, json } from '@sveltejs/kit';

import { getInternalHealth, isInternalHealthHost } from '$lib/server/health';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, url }) => {
	if (!isInternalHealthHost(url.hostname)) {
		throw error(404, 'Not found');
	}

	const { status, body } = await getInternalHealth(fetch);
	return json(body, { status });
};
