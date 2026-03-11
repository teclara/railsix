import { error, type Handle } from '@sveltejs/kit';

const RATE_LIMIT = 60; // max requests per window
const RATE_WINDOW_MS = 60_000; // 1 minute
const SSE_MAX_PER_IP = 3;

const hits = new Map<string, { count: number; resetAt: number }>();
const sseConns = new Map<string, number>();

// Clean up stale entries every 5 minutes
setInterval(() => {
	const now = Date.now();
	for (const [ip, entry] of hits) {
		if (now > entry.resetAt) hits.delete(ip);
	}
}, 300_000);

function getClientIp(event: Parameters<Handle>[0]['event']): string {
	return (
		event.request.headers.get('x-forwarded-for')?.split(',')[0]?.trim() || event.getClientAddress()
	);
}

function isRateLimited(ip: string): boolean {
	const now = Date.now();
	const entry = hits.get(ip);

	if (!entry || now > entry.resetAt) {
		hits.set(ip, { count: 1, resetAt: now + RATE_WINDOW_MS });
		return false;
	}

	entry.count++;
	return entry.count > RATE_LIMIT;
}

export const handle: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

	if (!pathname.startsWith('/api/')) {
		const response = await resolve(event);
		return addSecurityHeaders(response);
	}

	const ip = getClientIp(event);

	// SSE: limit concurrent connections per IP (no token check — EventSource can't send headers)
	if (pathname === '/api/sse') {
		const current = sseConns.get(ip) ?? 0;
		if (current >= SSE_MAX_PER_IP) {
			throw error(429, 'Too many SSE connections');
		}
		sseConns.set(ip, current + 1);
		const response = await resolve(event);

		// Wrap the body to detect close without consuming it
		const original = response.body;
		if (original) {
			const { readable, writable } = new TransformStream();
			original
				.pipeTo(writable)
				.catch(() => {})
				.finally(() => {
					const count = (sseConns.get(ip) ?? 1) - 1;
					if (count <= 0) sseConns.delete(ip);
					else sseConns.set(ip, count);
				});
			return addSecurityHeaders(
				new Response(readable, {
					status: response.status,
					headers: response.headers
				})
			);
		}
		return addSecurityHeaders(response);
	}

	// API endpoints: require token + rate limit
	if (event.request.headers.get('X-Requested-With') !== 'de479e2f71a8527f93608d266fcfa32c') {
		throw error(403, 'Forbidden');
	}

	if (isRateLimited(ip)) {
		throw error(429, 'Too many requests');
	}

	const response = await resolve(event);
	return addSecurityHeaders(response);
};

function addSecurityHeaders(response: Response): Response {
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');
	return response;
}
