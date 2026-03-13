import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('$app/environment', () => ({
	dev: false
}));

vi.mock('$env/dynamic/private', () => ({
	env: {
		API_BASE_URL: 'http://departures-api.railway.internal:8080',
		SSE_PUSH_URL: 'http://sse-push.railway.internal:8080',
		RAILWAY_PRIVATE_DOMAIN: 'web.railway.internal'
	}
}));

import { getInternalHealth, getPublicHealth, isInternalHealthHost } from './health';

describe('web health checks', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('returns ok for the public liveness endpoint without dependency details', async () => {
		await expect(getPublicHealth()).resolves.toEqual({
			status: 200,
			body: {
				status: 'ok'
			}
		});
	});

	it('returns ok when both backend dependencies are healthy', async () => {
		const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 200 }));

		await expect(getInternalHealth(fetchMock as typeof fetch)).resolves.toEqual({
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
		expect(fetchMock.mock.calls[0]?.[0]).toBe('http://departures-api.railway.internal:8080/ready');
		expect(fetchMock.mock.calls[1]?.[0]).toBe('http://sse-push.railway.internal:8080/ready');
	});

	it('returns service unavailable when a dependency is unhealthy', async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValueOnce(new Response(null, { status: 200 }))
			.mockResolvedValueOnce(new Response('unhealthy', { status: 503 }));

		await expect(getInternalHealth(fetchMock as typeof fetch)).resolves.toEqual({
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

	it('only allows the detailed health endpoint on internal hosts', () => {
		expect(isInternalHealthHost('web.railway.internal')).toBe(true);
		expect(isInternalHealthHost('custom.railway.internal')).toBe(true);
		expect(isInternalHealthHost('railsix.com')).toBe(false);
	});
});
