import { getSseUrl } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ request }) => {
	const sseUrl = getSseUrl();
	if (!sseUrl) {
		return new Response('SSE not configured', { status: 503 });
	}

	// Forward the client's abort signal so the upstream connection is cleaned up
	// when the browser disconnects.
	const upstream = await fetch(`${sseUrl}/sse`, { signal: request.signal });
	if (!upstream.ok || !upstream.body) {
		return new Response('SSE upstream unavailable', { status: 502 });
	}

	return new Response(upstream.body, {
		headers: {
			'Content-Type': 'text/event-stream',
			'Cache-Control': 'no-cache',
			Connection: 'keep-alive',
			'X-Accel-Buffering': 'no'
		}
	});
};
