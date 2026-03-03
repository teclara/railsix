// web/src/routes/schedule/+page.server.ts
import { getScheduleLines } from '$lib/api';

export async function load({ url }) {
	const date = url.searchParams.get('date') || new Date().toISOString().split('T')[0];
	try {
		const lines = await getScheduleLines(date);
		return { lines: Array.isArray(lines) ? lines : [], date };
	} catch {
		return { lines: [], date };
	}
}
