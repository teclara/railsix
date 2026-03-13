import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import { error } from '@sveltejs/kit';

const cacheHeaders = ['cache-control', 'content-type', 'etag', 'expires', 'last-modified', 'vary'];

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

export function withUpstreamTimeout(
	signal: AbortSignal | undefined,
	timeoutMs: number
): AbortSignal {
	const timeout = AbortSignal.timeout(timeoutMs);
	return signal ? AbortSignal.any([signal, timeout]) : timeout;
}

export async function proxyFetch(path: string, signal?: AbortSignal): Promise<Response> {
	const res = await fetch(`${getBaseUrl()}${path}`, {
		signal: withUpstreamTimeout(signal, 10000)
	});
	if (!res.ok) {
		throw error(mapProxyStatus(res.status), proxyErrorMessage(res.status));
	}

	const headers = new Headers();
	for (const name of cacheHeaders) {
		const value = res.headers.get(name);
		if (value) {
			headers.set(name, value);
		}
	}

	return new Response(res.body, {
		status: res.status,
		headers
	});
}

function mapProxyStatus(status: number): number {
	if (status >= 500) {
		return 502;
	}
	return status;
}

function proxyErrorMessage(status: number): string {
	if (status === 400) {
		return 'Invalid request';
	}
	if (status === 404) {
		return 'Resource not found';
	}
	if (status >= 500) {
		return 'Upstream service unavailable';
	}
	return 'Request rejected';
}
