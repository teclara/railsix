import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

import { getAllStops } from '$lib/api';
import { findStopBySlug } from '$lib/stations';

export const load: PageServerLoad = async ({ params }) => {
	const slug = params.station;

	if (!/^[a-z0-9-]+$/.test(slug)) {
		throw error(400, 'Invalid station');
	}

	try {
		const stops = await getAllStops();
		if (!Array.isArray(stops)) {
			throw error(502, 'Invalid response from departures-api');
		}

		const matched = findStopBySlug(stops, slug);
		if (!matched) {
			redirect(302, '/departures/union');
		}

		return {
			stops,
			stationCode: matched.code || matched.id,
			stationSlug: slug
		};
	} catch (err) {
		if (err instanceof Object && 'status' in err) throw err;
		console.error('[SSR] Failed to load departures page data:', err);
		throw error(503, 'Unable to load departures page data');
	}
};
