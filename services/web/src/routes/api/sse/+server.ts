import { getSseUrl, withUpstreamTimeout } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ request }) => {
	const sseUrl = getSseUrl();

	let upstream: Response;
	try {
		upstream = await fetch(`${sseUrl}/sse`, {
			signal: withUpstreamTimeout(request.signal, 10000)
		});
	} catch {
		return new Response('SSE upstream unreachable', { status: 502 });
	}
	if (!upstream.ok || !upstream.body) {
		return new Response('SSE upstream unavailable', { status: 502 });
	}

	request.signal.addEventListener(
		'abort',
		() => {
			void upstream.body?.cancel();
		},
		{ once: true }
	);

	return new Response(upstream.body, {
		headers: {
			'Content-Type': 'text/event-stream',
			'Cache-Control': 'no-cache',
			'X-Accel-Buffering': 'no'
		}
	});
};
