import { json } from '@sveltejs/kit';
import { getUnionDepartures } from '$lib/api';

export async function GET() {
	try {
		const departures = await getUnionDepartures();
		return json(departures);
	} catch {
		return json([], { status: 502 });
	}
}
