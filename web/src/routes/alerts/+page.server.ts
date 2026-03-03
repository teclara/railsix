// web/src/routes/alerts/+page.server.ts
import { getServiceAlerts, getInfoAlerts, getExceptions } from '$lib/api';

export async function load() {
	const [serviceAlerts, infoAlerts, exceptions] = await Promise.all([
		getServiceAlerts().catch(() => []),
		getInfoAlerts().catch(() => []),
		getExceptions().catch(() => [])
	]);

	return {
		serviceAlerts: Array.isArray(serviceAlerts) ? serviceAlerts : [],
		infoAlerts: Array.isArray(infoAlerts) ? infoAlerts : [],
		exceptions: Array.isArray(exceptions) ? exceptions : []
	};
}
