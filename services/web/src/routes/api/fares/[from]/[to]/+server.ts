import { proxyFetch } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = ({ params }) => {
	const from = encodeURIComponent(params.from);
	const to = encodeURIComponent(params.to);
	return proxyFetch(`/api/fares/${from}/${to}`);
};
