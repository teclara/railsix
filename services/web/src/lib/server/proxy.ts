import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import { error } from '@sveltejs/kit';

function getBaseUrl(): string {
	const url = env.API_BASE_URL || (dev ? 'http://localhost:8082' : '');
	if (!url) {
		throw error(500, 'API_BASE_URL environment variable is required in production');
	}
	return url;
}

export function getSseUrl(): string {
	const url = env.SSE_PUSH_URL || (dev ? 'http://localhost:8085' : '');
	if (!url) {
		throw error(500, 'SSE_PUSH_URL environment variable is required in production');
	}
	return url;
}

export async function proxyFetch(path: string): Promise<Response> {
	let res: globalThis.Response;
	try {
		res = await fetch(`${getBaseUrl()}${path}`, {
			signal: AbortSignal.timeout(10000)
		});
	} catch (err) {
		if (err instanceof DOMException && err.name === 'TimeoutError') {
			throw error(504, 'upstream timeout');
		}
		throw error(502, 'upstream unavailable');
	}
	if (!res.ok) {
		throw error(res.status, await res.text());
	}
	return new Response(res.body, {
		status: res.status,
		headers: { 'Content-Type': 'application/json' }
	});
}
