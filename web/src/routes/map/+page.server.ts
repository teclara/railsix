import { getPositions } from '$lib/api';

export async function load() {
	try {
		const positions = await getPositions();
		return { positions: Array.isArray(positions) ? positions : [] };
	} catch {
		return { positions: [] };
	}
}
