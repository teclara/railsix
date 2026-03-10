import { env } from '$env/dynamic/public';
import type { Alert } from './api';

/** API gateway base URL — set PUBLIC_API_URL in env (defaults to localhost:8080 for dev). */
function apiBase(): string {
	return env.PUBLIC_API_URL || 'http://localhost:8080';
}

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
		this.name = 'ApiError';
	}
}

export async function fetchAlerts(): Promise<Alert[]> {
	const res = await fetch(`${apiBase()}/api/alerts`, { signal: AbortSignal.timeout(10000) });
	if (!res.ok) throw new ApiError(res.status, `alerts: ${res.status}`);
	return res.json();
}

export type Departure = {
	line: string;
	lineName?: string;
	scheduledTime: string;
	actualTime?: string;
	arrivalTime?: string;
	status: string;
	platform?: string;
	delayMinutes?: number;
	stops?: string[];
	cars?: string;
	isInMotion?: boolean;
	isCancelled?: boolean;
	isExpress?: boolean;
	alert?: string;
	routeType?: number;
};

export async function fetchDepartures(stopCode: string, destCode?: string): Promise<Departure[]> {
	const base = apiBase();
	const url = destCode
		? `${base}/api/departures/${encodeURIComponent(stopCode)}?dest=${encodeURIComponent(destCode)}`
		: `${base}/api/departures/${encodeURIComponent(stopCode)}`;
	const res = await fetch(url, { signal: AbortSignal.timeout(10000) });
	if (!res.ok) throw new ApiError(res.status, `departures: ${res.status}`);
	return res.json();
}

export type UnionDeparture = {
	service: string;
	platform: string;
	time: string;
	info: string;
	stops: string[];
	cars?: string;
	isInMotion?: boolean;
	isCancelled?: boolean;
	alert?: string;
};

export async function fetchUnionDepartures(): Promise<UnionDeparture[]> {
	const res = await fetch(`${apiBase()}/api/union-departures`, {
		signal: AbortSignal.timeout(10000)
	});
	if (!res.ok) throw new ApiError(res.status, `union-departures: ${res.status}`);
	return res.json();
}

export type NetworkLine = {
	lineCode: string;
	lineName: string;
	activeTrips: number;
};

export async function fetchNetworkHealth(): Promise<NetworkLine[]> {
	const res = await fetch(`${apiBase()}/api/network-health`, {
		signal: AbortSignal.timeout(10000)
	});
	if (!res.ok) throw new ApiError(res.status, `network-health: ${res.status}`);
	return res.json();
}

export type FareInfo = {
	category: string;
	fareType: string;
	amount: number;
};

export async function fetchFares(from: string, to: string): Promise<FareInfo[]> {
	const res = await fetch(
		`${apiBase()}/api/fares/${encodeURIComponent(from)}/${encodeURIComponent(to)}`,
		{
			signal: AbortSignal.timeout(10000)
		}
	);
	if (!res.ok) throw new ApiError(res.status, `fares: ${res.status}`);
	return res.json();
}
