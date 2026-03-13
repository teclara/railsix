import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('$app/environment', () => ({
	dev: false
}));

vi.mock('$env/dynamic/private', () => ({
	env: {
		API_BASE_URL: 'http://departures-api.railway.internal:8080',
		SSE_PUSH_URL: 'http://sse-push.railway.internal:8080'
	}
}));

import { proxyFetch } from './proxy';

describe('proxyFetch', () => {
	let fetchMock: ReturnType<typeof vi.fn>;

	beforeEach(() => {
		fetchMock = vi.fn();
		vi.stubGlobal('fetch', fetchMock);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
	});

	it('forwards selected cache headers on success', async () => {
		fetchMock.mockResolvedValue(
			new Response('{"ok":true}', {
				status: 200,
				headers: {
					'Content-Type': 'application/json',
					'Cache-Control': 'public, max-age=30',
					Etag: '"abc123"'
				}
			})
		);

		const res = await proxyFetch('/alerts');

		expect(res.status).toBe(200);
		expect(res.headers.get('content-type')).toBe('application/json');
		expect(res.headers.get('cache-control')).toBe('public, max-age=30');
		expect(res.headers.get('etag')).toBe('"abc123"');
	});

	it('sanitizes upstream 5xx responses', async () => {
		fetchMock.mockResolvedValue(new Response('redis timeout', { status: 503 }));

		await expect(proxyFetch('/alerts')).rejects.toMatchObject({
			status: 502,
			body: { message: 'Upstream service unavailable' }
		});
	});
});
