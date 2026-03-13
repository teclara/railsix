import { dev } from '$app/environment';
import { env } from '$env/dynamic/private';

type DependencyStatus = 'ok' | 'error';

interface DependencyCheck {
	status: DependencyStatus;
	message?: string;
}

export interface WebHealthResponse {
	status: DependencyStatus;
	checks?: Record<string, DependencyCheck>;
}

function getApiBaseUrl() {
	const url = env.API_BASE_URL || (dev ? 'http://localhost:8082' : '');
	if (!url) {
		return null;
	}
	return url.replace(/\/+$/, '');
}

function getSseBaseUrl() {
	const url = env.SSE_PUSH_URL || (dev ? 'http://localhost:8085' : '');
	if (!url) {
		return null;
	}
	return url.replace(/\/+$/, '');
}

function describeError(err: unknown) {
	return err instanceof Error ? err.message : 'request failed';
}

async function checkDependency(
	fetchImpl: typeof fetch,
	name: string,
	baseUrl: string | null
): Promise<[string, DependencyCheck]> {
	if (!baseUrl) {
		return [name, { status: 'error', message: 'service URL is not configured' }];
	}

	try {
		const response = await fetchImpl(`${baseUrl}/ready`, {
			signal: AbortSignal.timeout(3000)
		});
		if (!response.ok) {
			return [
				name,
				{
					status: 'error',
					message: `unexpected status ${response.status}`
				}
			];
		}
		return [name, { status: 'ok' }];
	} catch (err) {
		return [name, { status: 'error', message: describeError(err) }];
	}
}

export async function getPublicHealth(): Promise<{ status: number; body: WebHealthResponse }> {
	return {
		status: 200,
		body: {
			status: 'ok'
		}
	};
}

export async function getInternalHealth(
	fetchImpl: typeof fetch = fetch
): Promise<{ status: number; body: WebHealthResponse }> {
	const results = await Promise.all([
		checkDependency(fetchImpl, 'api', getApiBaseUrl()),
		checkDependency(fetchImpl, 'ssePush', getSseBaseUrl())
	]);

	const checks = Object.fromEntries(results);
	const status = Object.values(checks).every((check) => check.status === 'ok') ? 'ok' : 'error';

	return {
		status: status === 'ok' ? 200 : 503,
		body: {
			status,
			checks
		}
	};
}

export function isInternalHealthHost(hostname: string): boolean {
	if (dev && (hostname === 'localhost' || hostname === '127.0.0.1')) {
		return true;
	}

	const privateDomain = env.RAILWAY_PRIVATE_DOMAIN?.trim();
	if (privateDomain && hostname === privateDomain) {
		return true;
	}

	return hostname.endsWith('.railway.internal');
}
