import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('$lib/api', () => ({
	getAllStops: vi.fn()
}));

import { getAllStops } from '$lib/api';
import { load } from './+page.server';

describe('departures server load', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('returns stops when departures page data loads successfully', async () => {
		vi.mocked(getAllStops).mockResolvedValue([{ code: 'UN', id: '1', name: 'Union' }]);

		await expect(load()).resolves.toEqual({
			stops: [{ code: 'UN', id: '1', name: 'Union' }]
		});
	});

	it('throws a 503 when departures page data cannot be loaded', async () => {
		vi.mocked(getAllStops).mockRejectedValue(new Error('backend unavailable'));

		await expect(load()).rejects.toMatchObject({ status: 503 });
	});
});
