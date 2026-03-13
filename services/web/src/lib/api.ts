import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import type { Alert, DeparturesResult, NetworkLine, Stop, UnionDeparture } from './api-contract';

function getApiBaseUrl() {
	const url = env.API_BASE_URL || (dev ? 'http://localhost:8082' : '');
	if (!url) {
		throw new Error('API_BASE_URL environment variable is required in production');
	}
	return url;
}

async function fetchApi<T>(baseUrl: string, path: string): Promise<T> {
	const res = await fetch(`${baseUrl}${path}`, {
		signal: AbortSignal.timeout(5000)
	});
	if (!res.ok) {
		throw new Error(`API error: ${res.status} ${res.statusText}`);
	}
	return res.json();
}

export function getAllStops() {
	return fetchApi<Stop[]>(getApiBaseUrl(), '/stops');
}

export function getStopDepartures(stopCode: string, destCode?: string) {
	const path = destCode
		? `/departures/${encodeURIComponent(stopCode)}?dest=${encodeURIComponent(destCode)}`
		: `/departures/${encodeURIComponent(stopCode)}`;
	return fetchApi<DeparturesResult>(getApiBaseUrl(), path);
}

export function getAlerts() {
	return fetchApi<Alert[]>(getApiBaseUrl(), '/alerts');
}

export function getUnionDepartures() {
	return fetchApi<UnionDeparture[]>(getApiBaseUrl(), '/union-departures');
}

export function getNetworkHealth() {
	return fetchApi<NetworkLine[]>(getApiBaseUrl(), '/network-health');
}

export type { Alert, DeparturesResult, NetworkLine, Stop, UnionDeparture } from './api-contract';
