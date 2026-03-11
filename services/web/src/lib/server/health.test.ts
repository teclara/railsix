import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('$app/environment', () => ({
	dev: false
}));

vi.mock('$env/dynamic/private', () => ({
	env: {
		API_BASE_URL: 'http://departures-api.railway.internal:8080',
		SSE_PUSH_URL: 'http://sse-push.railway.internal:8080'
	}
}));

import { getWebHealth } from './health';

describe('web health checks', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('returns ok when both backend dependencies are healthy', async () => {
		const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 200 }));

		await expect(getWebHealth(fetchMock as typeof fetch)).resolves.toEqual({
			status: 200,
			body: {
				status: 'ok',
				checks: {
					api: { status: 'ok' },
					ssePush: { status: 'ok' }
				}
			}
		});

		expect(fetchMock).toHaveBeenCalledTimes(2);
	});

	it('returns service unavailable when a dependency is unhealthy', async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValueOnce(new Response(null, { status: 200 }))
			.mockResolvedValueOnce(new Response('unhealthy', { status: 503 }));

		await expect(getWebHealth(fetchMock as typeof fetch)).resolves.toEqual({
			status: 503,
			body: {
				status: 'error',
				checks: {
					api: { status: 'ok' },
					ssePush: { status: 'error', message: 'unexpected status 503' }
				}
			}
		});
	});
});
