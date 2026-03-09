import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getFares } from '$lib/api';

const stopCodeRe = /^[A-Za-z0-9]{2,10}$/;

export const GET: RequestHandler = async ({ params }) => {
	if (!stopCodeRe.test(params.from) || !stopCodeRe.test(params.to)) {
		return json({ error: 'invalid stop code' }, { status: 400 });
	}
	try {
		const fares = await getFares(params.from, params.to);
		return json(fares);
	} catch (err) {
		console.error('[proxy] /api/fares failed:', err);
		return json({ error: 'upstream unavailable' }, { status: 502 });
	}
};
