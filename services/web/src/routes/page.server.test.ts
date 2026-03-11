import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('$lib/api', () => ({
	getAllStops: vi.fn(),
	getAlerts: vi.fn()
}));

import { getAlerts, getAllStops } from '$lib/api';
import { load } from './+page.server';

describe('homepage server load', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('returns SSR data when both backend calls succeed', async () => {
		vi.mocked(getAllStops).mockResolvedValue([{ code: 'UN', id: '1', name: 'Union' }]);
		vi.mocked(getAlerts).mockResolvedValue([{ headline: 'Notice', description: 'Normal service' }]);

		await expect(load()).resolves.toEqual({
			stops: [{ code: 'UN', id: '1', name: 'Union' }],
			alerts: [{ headline: 'Notice', description: 'Normal service' }]
		});
	});

	it('throws a 503 when homepage data cannot be loaded', async () => {
		vi.mocked(getAllStops).mockRejectedValue(new Error('backend unavailable'));
		vi.mocked(getAlerts).mockResolvedValue([]);

		await expect(load()).rejects.toMatchObject({ status: 503 });
	});
});
