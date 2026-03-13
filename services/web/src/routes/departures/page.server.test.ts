import { describe, expect, it } from 'vitest';

import { load } from './+page.server';

describe('departures redirect', () => {
	it('redirects /departures to /departures/union', () => {
		expect(() => load()).toThrow();
		try {
			load();
		} catch (e) {
			expect(e).toMatchObject({ status: 302, location: '/departures/union' });
		}
	});
});
