# Six Rail — World-Class Commuter Upgrade Design

**Date:** 2026-03-06
**Status:** Approved

## Overview

Upgrade Six Rail from a map-first transit tracker to a world-class commuter app. Primary audience: GO Transit daily commuters. Core positioning: free forever, no ads, no upsells — in a market where competitors (Transit App, Go Rider) charge or gatekeep features.

**Key market opportunity:** Triplinx shut down January 2025, leaving a gap for GO Transit commuters. The official Metrolinx app is mediocre. Six Rail can own this space.

## Goals

- Reduce friction to near zero for daily commuters checking their next train
- Provide richer departure info (platform, on-time status, delay in minutes)
- Make the app installable and feel native via PWA
- Deliver proactive delay notifications without a paid push service

## Non-Goals (V1)

- Journey planner / trip routing
- Ticketing or fare info
- Crowding data
- Server-side push notifications (V2)
- Native iOS / Android apps

## Architecture

No backend changes required. All new features are frontend-only, using existing API endpoints. The departure data already includes delay and route info; platform numbers are surfaced from GTFS trip updates where available.

## Information Architecture

Two top-level views via bottom nav bar (mobile) / side nav (desktop):

| Tab | Description |
|-----|-------------|
| Dashboard | Default PWA launch view — commute status, split-flap board, countdown |
| Map | Existing fullscreen Mapbox map (unchanged, minor additions) |

Settings accessed via gear icon (top-right of dashboard) — not a nav tab.

## Dashboard Design

### Layout

```
┌─────────────────────────────────┐
│  Good morning, [name]      ⚙️   │
│  Day · Date                     │
├─────────────────────────────────┤
│  [ TO WORK ]  ⇄  [ TO HOME ]   │  <- auto-selected by time of day
├─────────────────────────────────┤
│  ┌─── SPLIT-FLAP BOARD ──────┐  │
│  │ OAKVILLE → UNION STATION  │  │
│  │                           │  │
│  │ 08:15  Lakeshore East  ✦  │  │  <- next train (highlighted)
│  │        Platform 3         │  │
│  │        ON TIME            │  │
│  │                           │  │
│  │ 08:45  Lakeshore East     │  │
│  │        DELAYED  +8 min    │  │
│  │                           │  │
│  │ 09:15  Lakeshore East     │  │
│  │        ON TIME            │  │
│  └───────────────────────────┘  │
│                                 │
│  ┌─────────────────────────┐    │
│  │  Next train in  12:34   │    │  <- live countdown
│  └─────────────────────────┘    │
│                                 │
│  🔔 Notify me if delayed        │
└─────────────────────────────────┘
```

### Trip Direction Logic

- Before noon: auto-select "to work" trip
- After noon: auto-select "to home" trip
- User can tap toggle to manually override at any time

### Departure Rows

- Show 3 departures on mobile, 5 on desktop
- Next departure is larger/brighter — the "catch this one" train
- Each row: departure time, route name, platform (if available), status badge
- Status badges: `ON TIME` (green), `DELAYED +N min` (amber), `CANCELLED` (red strikethrough)

## Split-Flap Display Component

A custom Svelte component that simulates a Solari split-flap departure board.

**Visual spec:**
- Board background: `#111`
- Character tiles: `#1e1e1e` with subtle inset shadow, `2px` border-radius
- Text colors: `#f5a623` amber for times, `#ffffff` for names, `#ef4444` for DELAYED/CANCELLED, `#22c55e` for ON TIME
- Font: `JetBrains Mono` or `IBM Plex Mono` (monospace)
- Flip animation: CSS `rotateX` on card halves, 80ms per character, staggered left-to-right
- Characters cycle through intermediate values before landing on final value (authentic feel)
- Only characters that change trigger the flip animation on refresh

**Behaviour:**
- Refreshes every 30s (same interval as existing departures panel)
- Smooth, non-jarring updates — unchanged characters do not animate

## PWA

**Manifest (`manifest.json`):**
- `display: standalone` — no browser chrome when launched from home screen
- Theme color: `#111`
- Background color: `#111`
- App icon: train on dark background (192x192, 512x512)
- `start_url`: `/` (opens dashboard)

**Service Worker:**
- Registered on app load
- Caches static assets for offline fallback
- Polls departure API every 2 minutes in background
- Fires local Web Push notification if next train delay increases by more than the user's threshold (default: 5 min)

No server-side push infrastructure needed for V1. Notifications fire while the device has connectivity and the service worker is active.

## Notifications

**Trigger:** User taps "Notify me if delayed" on dashboard.
**Permission:** Standard browser `Notification.requestPermission()` flow.
**Storage:** Notification preference and threshold stored in localStorage.
**Logic:** Service worker compares cached departure delay vs. latest API response. If delta exceeds threshold, fire notification: "Your 08:15 Lakeshore East is now delayed 12 min."

## Alerts Integration

- Active alerts that match the user's saved commute routes surface as a slim amber banner above the split-flap board on the dashboard
- Format: "Delays on Lakeshore East — tap for details"
- Full alerts list remains accessible via the map tab (existing AlertsDropdown)

## First-Run Onboarding

Shown when no commute trips are saved. Inline on the dashboard — no separate route.

1. Empty state with CTA: "Set up your commute to get started"
2. Step 1: Pick origin and destination for "going to work"
3. Step 2: Pre-fills reverse for "going home", user confirms or adjusts
4. Stored in localStorage (same pattern as existing `favorites` store)

## Settings Screen

Accessed via gear icon on dashboard. Covers:
- Edit "to work" trip (origin, destination)
- Edit "to home" trip (origin, destination)
- Notification toggle + delay threshold (5 / 10 / 15 min)
- Clear all data

## Map Tab Changes

Minimal changes — the map is already good:
- Add "Jump to my station" floating button — flies to saved default station and opens departures panel
- Alerts dropdown stays for full alert list
- All existing features unchanged (filters, vehicle popups, route lines, search)

## New Stores (localStorage)

```typescript
// Commute trips
interface CommuteTrip {
  originCode: string;
  originName: string;
  destinationCode: string;
  destinationName: string;
}

interface CommuteStore {
  toWork: CommuteTrip | null;
  toHome: CommuteTrip | null;
}

// Notification preferences
interface NotificationStore {
  enabled: boolean;
  thresholdMinutes: 5 | 10 | 15;
}
```

## New Components

| Component | Description |
|-----------|-------------|
| `SplitFlapBoard.svelte` | Split-flap display with flip animation |
| `SplitFlapChar.svelte` | Individual character tile with CSS flip |
| `CommuteDashboard.svelte` | Dashboard page layout |
| `CommuteSetup.svelte` | First-run onboarding flow |
| `CountdownTimer.svelte` | Live countdown to next departure |
| `AlertBanner.svelte` | Slim alert banner for commute route alerts |
| `SettingsPanel.svelte` | Settings screen (gear icon) |

## New Routes

| Route | Description |
|-------|-------------|
| `/` | Dashboard (was map, now dashboard) |
| `/map` | Map (existing page moved) |

## Feature Summary

| Feature | Status |
|---------|--------|
| Split-flap commute dashboard | New |
| Auto time-aware trip direction | New |
| Live countdown to next train | New |
| Platform + on-time/delayed/cancelled status | New |
| PWA manifest + service worker | New |
| Local push notifications | New |
| Alert banner for saved commute routes | New |
| First-run onboarding | New |
| Settings screen | New |
| "Jump to my station" map button | New |
| Map, filters, vehicle popups, route lines | Existing |
