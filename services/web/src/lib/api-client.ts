import type { Alert, DeparturesResult, NetworkLine, UnionDeparture } from './api-contract';

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
		this.name = 'ApiError';
	}
}

export async function fetchAlerts(signal?: AbortSignal): Promise<Alert[]> {
	const res = await fetch('/api/alerts', {
		signal: signal ?? AbortSignal.timeout(10000)
	});
	if (!res.ok) throw new ApiError(res.status, `alerts: ${res.status}`);
	return res.json();
}

export async function fetchDepartures(
	stopCode: string,
	destCode?: string,
	signal?: AbortSignal
): Promise<DeparturesResult> {
	const url = destCode
		? `/api/departures/${encodeURIComponent(stopCode)}?dest=${encodeURIComponent(destCode)}`
		: `/api/departures/${encodeURIComponent(stopCode)}`;
	const res = await fetch(url, {
		signal: signal ?? AbortSignal.timeout(10000)
	});
	if (!res.ok) throw new ApiError(res.status, `departures: ${res.status}`);
	return res.json();
}

export async function fetchUnionDepartures(): Promise<UnionDeparture[]> {
	const res = await fetch('/api/union-departures', {
		signal: AbortSignal.timeout(10000)
	});
	if (!res.ok) throw new ApiError(res.status, `union-departures: ${res.status}`);
	return res.json();
}

export async function fetchNetworkHealth(): Promise<NetworkLine[]> {
	const res = await fetch('/api/network-health', {
		signal: AbortSignal.timeout(10000)
	});
	if (!res.ok) throw new ApiError(res.status, `network-health: ${res.status}`);
	return res.json();
}

export type {
	Alert,
	Departure,
	DeparturesResult,
	NetworkLine,
	UnionDeparture
} from './api-contract';
