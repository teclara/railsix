// web/src/lib/api.ts
import { env } from '$env/dynamic/private';

const baseUrl = env.API_BASE_URL || 'http://localhost:8080';

async function fetchApi<T>(path: string): Promise<T> {
	const res = await fetch(`${baseUrl}${path}`);
	if (!res.ok) {
		throw new Error(`API error: ${res.status} ${res.statusText}`);
	}
	return res.json();
}

export function getUnionDepartures() {
	return fetchApi('/api/departures/union');
}

export function getStopDepartures(stopCode: string) {
	return fetchApi(`/api/departures/${stopCode}`);
}

export function getTrains() {
	return fetchApi('/api/trains');
}

export function getTrainPositions() {
	return fetchApi('/api/trains/positions');
}

export function getServiceAlerts() {
	return fetchApi('/api/alerts/service');
}

export function getInfoAlerts() {
	return fetchApi('/api/alerts/info');
}

export function getExceptions() {
	return fetchApi('/api/exceptions');
}

export function getScheduleLines(date: string) {
	return fetchApi(`/api/schedule/lines/${date}`);
}

export function getScheduleJourney(params: {
	date: string;
	from: string;
	to: string;
	startTime: string;
	maxJourney?: string;
}) {
	const query = new URLSearchParams({
		date: params.date,
		from: params.from,
		to: params.to,
		startTime: params.startTime,
		maxJourney: params.maxJourney || '3'
	});
	return fetchApi(`/api/schedule/journey?${query}`);
}

export function getFares(from: string, to: string) {
	return fetchApi(`/api/fares/${from}/${to}`);
}

export function getAllStops() {
	return fetchApi('/api/stops');
}

export function getStopDetails(code: string) {
	return fetchApi(`/api/stops/${code}`);
}
