# Six Rail SEO Audit — March 2026

## Executive Summary

Six Rail is **not indexed by Google at all**. A `site:sixrail.up.railway.app` search returns zero results. The site has decent meta tags on the homepage but is missing meta descriptions on `/board`, has no `robots.txt`, no `sitemap.xml`, no structured data, and no content pages for informational queries. The competitive landscape is crowded but beatable — most competitors are small indie projects with thin SEO.

**Biggest strength:** Real-time, functional product with a clean UI that solves a real commuter need.

**Top 3 priorities:**
1. Get indexed — add robots.txt, sitemap.xml, submit to Google Search Console
2. Add meta descriptions and structured data to all pages
3. Create SEO landing pages for high-intent schedule queries (station-specific, line-specific)

**Overall assessment:** Critical issues — the site is functionally invisible to search engines.

---

## Keyword Opportunities

| Keyword | Est. Difficulty | Opportunity | Intent | Recommended Content |
|---------|----------------|-------------|--------|---------------------|
| go train schedule | Hard | High | Navigational | Homepage + line pages |
| go train schedule today | Moderate | High | Transactional | Dynamic schedule page |
| go transit departure times | Moderate | High | Transactional | Board page (with SEO) |
| next go train from union station | Easy | High | Transactional | Board / Union landing |
| go train tracker | Moderate | High | Navigational | Homepage |
| go transit real time | Moderate | High | Transactional | Homepage |
| union station go departures | Easy | High | Transactional | /board landing page |
| lakeshore east go train schedule | Easy | High | Informational | Line-specific page |
| lakeshore west go train schedule | Easy | High | Informational | Line-specific page |
| barrie go train schedule | Easy | High | Informational | Line-specific page |
| kitchener go train schedule | Easy | High | Informational | Line-specific page |
| go train delays today | Moderate | Medium | Informational | Alerts/status page |
| go transit service alerts | Moderate | Medium | Informational | Alerts page |
| go train departure board | Easy | Medium | Transactional | /board page |
| when is the next go train | Easy | Medium | Transactional | Homepage/board |
| go transit live train map | Hard | Medium | Navigational | Future feature page |
| go train platform info | Easy | Medium | Transactional | Board page |
| go train commute tracker | Easy | Medium | Commercial | Homepage |
| go transit oakville schedule | Easy | Medium | Informational | Station-specific page |
| go transit ajax train times | Easy | Medium | Informational | Station-specific page |
| go train cancelled today | Moderate | Medium | Informational | Alerts page |
| go transit fare calculator | Hard | Low | Commercial | Fare page |
| how often do go trains run | Easy | Low | Informational | FAQ/blog post |
| go train weekend schedule | Easy | Low | Informational | Schedule page |

---

## On-Page Issues

| Page | Issue | Severity | Fix |
|------|-------|----------|-----|
| `/board` | No meta description | Critical | Add keyword-rich meta description |
| `/board` | No Open Graph / Twitter tags | High | Add og:title, og:description, twitter:card |
| All pages | No robots.txt | Critical | Create `/static/robots.txt` |
| All pages | No sitemap.xml | Critical | Generate and serve `/sitemap.xml` |
| All pages | No canonical tags | High | Add `<link rel="canonical">` |
| All pages | No structured data (Schema.org) | High | Add WebApplication, BreadcrumbList schemas |
| `/` | H1 is "Six Rail" — no keywords | Medium | Change to "GO Transit Real-Time Train Tracker" |
| `/` | Title could be more keyword-dense | Medium | "GO Train Schedule & Real-Time Tracker — Six Rail" |
| `/board` | H1 is station name only | Medium | Include "GO Train Departures" in heading |
| All pages | No descriptive internal text links | Medium | Add body links between pages |
| `/` | 500+ station entries in HTML | Medium | Consider lazy-loading station data |

---

## Content Gaps

| Topic | Why It Matters | Format | Priority | Effort |
|-------|---------------|--------|----------|--------|
| Station-specific pages (`/stations/oakville`) | High-intent long-tail queries | Auto-generated landing pages | High | 2-3 days |
| Line-specific pages (`/lines/lakeshore-east`) | Common schedule queries | Auto-generated landing pages | High | 1-2 days |
| "GO Train Schedule Today" | Top search query | Dynamic page | High | Quick win |
| Alerts/delays page (`/alerts`) | "GO train delays today" queries | Dedicated page | Medium | Quick win |
| FAQ page | "How often do GO trains run" etc. | Static content | Medium | 1-2 hours |
| Blog: "How to Track Your GO Train" | Top-of-funnel traffic | Blog post | Low | Half day |

---

## Technical SEO Checklist

| Check | Status | Details |
|-------|--------|---------|
| HTTPS | Pass | Via Railway |
| Mobile viewport | Pass | Meta tag present |
| Page speed | Warning | Large station data payload in HTML |
| robots.txt | **Fail** | 404 — doesn't exist |
| sitemap.xml | **Fail** | 404 — doesn't exist |
| Canonical tags | **Fail** | None on any page |
| Structured data | **Fail** | No Schema.org markup |
| Google indexation | **Fail** | Zero pages indexed |
| PWA manifest | Pass | manifest.json + service worker |
| Open Graph | Warning | Only on `/`, missing on `/board` |
| HTML lang | Pass | `lang="en"` |
| Heading hierarchy | Warning | H1 is brand-only on homepage |
| Internal linking | Warning | Nav only, no body text links |
| Broken links | Pass | None detected |
| Mixed content | Pass | Clean HTTPS |

---

## Competitor Comparison

| Dimension | Six Rail | OnTime GO | trains.fyi | Toronto GO Tracker | gotrainschedule.com |
|-----------|---------|-----------|------------|-------------------|---------------------|
| Real-time departures | Yes | Yes | Map only | Yes | No (static) |
| Meta descriptions | Partial | Full | Full | Missing | Full |
| Structured data | None | Full | Partial | None | Partial |
| Google indexed | **No** | Yes | Yes | Yes | Yes |
| Content pages | 2 | Multiple | Multiple | 2 | 10+ |
| Station-specific URLs | No | Yes | No | Yes | Yes |
| Blog/SEO content | None | None | History page | None | Schedule guides |
| Mobile app | PWA | iOS+Android | iOS+Android | No | No |

Competitor URLs:
- https://www.ontimego.ca/
- https://trains.fyi/networks/go/
- https://torontogotracker.com/departures/union/
- https://gotrainschedule.com/

---

## Action Plan

### Quick Wins (do once domain is registered)
- Create `robots.txt` and `sitemap.xml`
- Submit to Google Search Console
- Add meta description + OG/Twitter tags to `/board`
- Add canonical tags to all pages
- Improve homepage H1 and title for keywords
- Register custom domain (e.g. `sixrail.ca`)

### Strategic Investments (this quarter)
- Auto-generate station-specific pages from GTFS data
- Auto-generate line-specific schedule pages
- Add Schema.org structured data (WebApplication, BreadcrumbList)
- Create dedicated alerts/status page
- Build FAQ page for long-tail queries
- Consider blog for informational content
