# Rail Six SEO Audit — March 2026 (Updated)

## Executive Summary

Rail Six (railsix.com) is **still not indexed by Google**. A `site:railsix.com` search returns zero results. Since the last audit, several improvements have been made — the domain is now a proper `.com`, the homepage and `/departures` page have meta descriptions and OG/Twitter tags, and a third page (`/departures`) has been added. However, **critical blockers remain**: no `robots.txt`, no `sitemap.xml`, no canonical tags, no structured data, and `/board` still has no meta description or social tags.

**Biggest strength:** Real-time, functional product with three distinct views (commute dashboard, station departures, split-flap board) and a clean UI that solves a real commuter need. Custom domain (railsix.com) is now live.

**What's improved since last audit:**
- Custom domain registered and live (railsix.com)
- Meta description added to homepage
- OG and Twitter tags added to `/` and `/departures`
- New `/departures` page with station search (third indexable page)
- Brand renamed from "Six Rail" to "Rail Six"

**Top 3 priorities:**
1. **Get indexed** — add `robots.txt`, `sitemap.xml`, submit to Google Search Console immediately
2. **Fix `/board` SEO** — add meta description, OG tags, canonical tag
3. **Add structured data and canonical tags** to all pages

**Overall assessment:** Critical — the site is functionally invisible to search engines despite having a proper domain.

---

## Keyword Opportunities

| Keyword | Est. Difficulty | Opportunity | Intent | Recommended Content |
|---------|----------------|-------------|--------|---------------------|
| go train schedule | Hard | High | Navigational | Homepage + line pages |
| go train schedule today | Moderate | High | Transactional | Dynamic schedule page |
| go transit departure times | Moderate | High | Transactional | `/departures` page |
| next go train from union station | Easy | High | Transactional | `/board` + Union landing |
| go train tracker | Moderate | High | Navigational | Homepage |
| go transit real time | Moderate | High | Transactional | Homepage |
| union station go departures | Easy | High | Transactional | `/board` landing page |
| go train departure board | Easy | High | Transactional | `/board` page |
| lakeshore east go train schedule | Easy | High | Informational | Line-specific page |
| lakeshore west go train schedule | Easy | High | Informational | Line-specific page |
| barrie go train schedule | Easy | High | Informational | Line-specific page |
| kitchener go train schedule | Easy | High | Informational | Line-specific page |
| stouffville go train schedule | Easy | High | Informational | Line-specific page |
| go train delays today | Moderate | Medium | Informational | Alerts/status page |
| go transit service alerts | Moderate | Medium | Informational | Alerts page |
| when is the next go train | Easy | Medium | Transactional | Homepage / `/departures` |
| go transit live train map | Hard | Medium | Navigational | Future feature page |
| go train platform info | Easy | Medium | Transactional | `/board` page |
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
| All pages | No `robots.txt` | **Critical** | Create `web/static/robots.txt` with `Sitemap: https://railsix.com/sitemap.xml` |
| All pages | No `sitemap.xml` | **Critical** | Generate and serve `/sitemap.xml` listing `/`, `/departures`, `/board` |
| All pages | No canonical tags | **Critical** | Add `<link rel="canonical" href="https://railsix.com{path}">` to every page |
| All pages | No structured data | **High** | Add `WebApplication` + `BreadcrumbList` JSON-LD schemas |
| `/board` | No meta description | **High** | Add: "Live GO Transit departure board — real-time train times, platforms, and delays from Union Station and all GO stations." |
| `/board` | No Open Graph / Twitter tags | **High** | Add `og:title`, `og:description`, `og:type`, `twitter:card`, `twitter:title`, `twitter:description` |
| `/board` | No `og:url` or `og:image` | **High** | Add og:url and a social share image |
| `/` | No `og:url` or `og:image` | **High** | Add `og:url="https://railsix.com"` and a social share image |
| `/departures` | No `og:url` or `og:image` | **High** | Add og:url and social share image |
| `/departures` | No Twitter card tags | **Medium** | Add `twitter:card`, `twitter:title`, `twitter:description` |
| `/` | H1 is "Rail Six" — no keywords | **Medium** | Change to "GO Transit Real-Time Train Tracker" or similar |
| `/` | Title could be more keyword-dense | **Medium** | Consider: "GO Train Schedule & Real-Time Tracker — Rail Six" |
| `/board` | H1 is station name only | **Medium** | Include "GO Departures" in the H1 (e.g. "Union Station GO Departures") |
| All pages | No descriptive body text links | **Medium** | Add contextual links between pages in body content |
| All pages | Large station data in SSR HTML | **Low** | Consider lazy-loading station lists to reduce payload |

---

## Content Gaps

| Topic | Why It Matters | Format | Priority | Effort |
|-------|---------------|--------|----------|--------|
| Station-specific pages (`/stations/oakville`) | High-intent long-tail queries like "GO transit oakville schedule" | Auto-generated landing pages with live departures | High | 2-3 days |
| Line-specific pages (`/lines/lakeshore-east`) | Common schedule queries like "lakeshore east go train schedule" | Auto-generated pages with route map + schedule | High | 1-2 days |
| "GO Train Schedule Today" | Top search query, highly transactional | Dynamic page showing today's full schedule | High | Quick win |
| Alerts/delays page (`/alerts`) | "GO train delays today" queries | Dedicated page pulling from existing alerts API | Medium | Quick win |
| FAQ page | "How often do GO trains run", "what platform", etc. | Static content with FAQ Schema markup | Medium | 1-2 hours |
| Blog: "How to Track Your GO Train in 2026" | Top-of-funnel traffic + backlink bait | Blog post | Low | Half day |

---

## Technical SEO Checklist

| Check | Status | Details |
|-------|--------|---------|
| HTTPS | **Pass** | Via Railway |
| Custom domain | **Pass** ✨ | `railsix.com` — improvement from previous `sixrail.up.railway.app` |
| Mobile viewport | **Pass** | `<meta name="viewport">` present |
| PWA manifest | **Pass** | `manifest.json` + service worker registered |
| HTML lang | **Pass** | `lang="en"` set |
| Broken links | **Pass** | None detected |
| Mixed content | **Pass** | Clean HTTPS |
| Homepage meta description | **Pass** ✨ | Added since last audit |
| Homepage OG/Twitter tags | **Pass** ✨ | Added since last audit |
| `/departures` meta + OG | **Pass** ✨ | New page with proper tags |
| Page speed | **Warning** | Large station data payload serialized in SSR HTML |
| Open Graph | **Warning** | Present on `/` and `/departures`, missing on `/board`. No `og:image` on any page |
| Heading hierarchy | **Warning** | H1 is brand-only on homepage; H1 on `/board` lacks keywords |
| Internal linking | **Warning** | Nav-only links, no contextual body text links |
| robots.txt | **Fail** | 404 — doesn't exist |
| sitemap.xml | **Fail** | 404 — doesn't exist |
| Canonical tags | **Fail** | None on any page |
| Structured data | **Fail** | No Schema.org markup (competitors OnTime GO and trains.fyi both have it) |
| Google indexation | **Fail** | Zero pages indexed (`site:railsix.com` returns nothing) |

---

## Competitor Comparison

| Dimension | Rail Six | OnTime GO | trains.fyi | GoTrack App | Go Rider |
|-----------|---------|-----------|------------|-------------|----------|
| Real-time departures | Yes | Yes | Map only | Yes | Yes |
| Custom domain | **Yes** ✨ | Yes | Yes | N/A (app) | N/A (app) |
| Meta descriptions | Partial (2/3 pages) | Full | Full | N/A | N/A |
| Structured data | **None** | Full (MobileApplication + WebSite) | Partial (WebSite) | N/A | N/A |
| Google indexed | **No** | Yes | Yes | App Store | App Store |
| Content pages | 3 | Multiple | Multiple (cities, networks, stats) | N/A | N/A |
| Station-specific URLs | No | Yes | No | N/A | N/A |
| Blog/SEO content | None | None | History + stats pages | N/A | N/A |
| Platform | PWA (web) | iOS + Android + Web | iOS + Android + Web | iOS | iOS |
| Fare info | Yes (API exists) | Yes (PRESTO fares) | No | No | No |
| Countdown timers | Yes | Yes | No | No | No |
| Service alerts | Yes | Yes | No | No | Yes |

Competitor URLs:
- https://www.ontimego.ca/
- https://trains.fyi/networks/go/
- https://apps.apple.com/ca/app/gotrack/id1265458198
- https://apps.apple.com/us/app/go-rider-train-schedules/id6502670635

**Key competitive insight:** OnTime GO is the strongest web competitor — they have full structured data (MobileApplication schema), keyword-rich title and meta description, and App Store presence. However, they are app-first. Rail Six's PWA approach means the **web is your distribution channel** — SEO is not optional, it's existential. You must out-SEO every competitor on the web to compensate for not being in App Store search.

---

## Action Plan

### Quick Wins (this week)

| Action | Impact | Effort |
|--------|--------|--------|
| Create `web/static/robots.txt` (`User-agent: *\nAllow: /\nSitemap: https://railsix.com/sitemap.xml`) | High | 5 min |
| Create and serve `sitemap.xml` listing `/`, `/departures`, `/board` | High | 30 min |
| Submit railsix.com to Google Search Console + request indexing | High | 15 min |
| Add meta description + OG + Twitter tags to `/board` | High | 15 min |
| Add `<link rel="canonical">` to all 3 pages | High | 15 min |
| Add `og:url` and `og:image` to all pages (create a social share image) | Medium | 1 hour |
| Add `twitter:card` tags to `/departures` page | Medium | 5 min |
| Submit sitemap URL in Google Search Console | High | 5 min |

### Strategic Investments (this quarter)

| Action | Impact | Effort |
|--------|--------|--------|
| Add Schema.org structured data — `WebApplication` on homepage, `BreadcrumbList` on all pages | High | Half day |
| Auto-generate station-specific pages from GTFS stop data (`/stations/{slug}`) | High | 2-3 days |
| Auto-generate line-specific pages (`/lines/{slug}`) | High | 1-2 days |
| Create dedicated `/alerts` page for "GO train delays today" queries | Medium | Quick win |
| Build FAQ page with `FAQPage` schema for long-tail queries | Medium | 1-2 hours |
| Improve homepage H1 to include keywords (e.g. "GO Transit Real-Time Train Tracker") | Medium | 5 min |
| Create social share preview image (OG image) | Medium | 1 hour |
| Bing Webmaster Tools submission | Low | 15 min |
| Consider blog content for informational queries | Low | Ongoing |

---

## Progress Since Last Audit

| Item from Original Action Plan | Status |
|-------------------------------|--------|
| Register custom domain | **Done** ✅ — railsix.com |
| Add meta description to homepage | **Done** ✅ |
| Add OG/Twitter tags to homepage | **Done** ✅ |
| Add `/departures` page | **Done** ✅ — new page with full SEO tags |
| Create `robots.txt` | Not done |
| Create `sitemap.xml` | Not done |
| Submit to Google Search Console | Not done |
| Add meta description to `/board` | Not done |
| Add canonical tags | Not done |
| Add structured data | Not done |
| Station-specific pages | Not done |
| Line-specific pages | Not done |
| Alerts page | Not done |
