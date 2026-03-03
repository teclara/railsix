// web/src/routes/+layout.server.ts
import { getServiceAlerts } from '$lib/api';

export async function load() {
	try {
		const alerts = await getServiceAlerts();
		return { alerts: Array.isArray(alerts) ? alerts : [] };
	} catch {
		return { alerts: [] };
	}
}
