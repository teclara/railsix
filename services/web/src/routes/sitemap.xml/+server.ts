import type { RequestHandler } from './$types';

import { getAllStops } from '$lib/api';
import { stopToSlug } from '$lib/stations';

export const GET: RequestHandler = async () => {
	const BASE = 'https://railsix.com';

	let stationUrls = '';
	try {
		const stops = await getAllStops();
		if (Array.isArray(stops)) {
			const slugs = new Set<string>();
			for (const stop of stops) {
				const slug = stopToSlug(stop);
				if (slug && !slugs.has(slug)) {
					slugs.add(slug);
					stationUrls += `
  <url>
    <loc>${BASE}/departures/${slug}</loc>
    <changefreq>daily</changefreq>
    <priority>0.7</priority>
  </url>`;
				}
			}
		}
	} catch {
		// If API is down, return sitemap without station URLs
	}

	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>${BASE}/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>${BASE}/departures</loc>
    <changefreq>daily</changefreq>
    <priority>0.9</priority>
  </url>${stationUrls}
</urlset>`;

	return new Response(xml, {
		headers: {
			'Content-Type': 'application/xml',
			'Cache-Control': 'max-age=3600'
		}
	});
};
