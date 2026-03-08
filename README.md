# Rail Six

Real-time GO Transit tracking for Toronto commuters. Split-flap departure board, commute dashboard with countdown timer, delay alerts, and network health — free, no ads.

## Features

- **Commute dashboard** — Save your to-work and to-home trips, see next departures with countdown timer
- **Split-flap departure board** — Full-screen board for Union Station or any GO station, scales for TV/kiosk display
- **Station lookup** — Search any GO station and view upcoming train departures
- **Real-time data** — Platform assignments, delay status, car counts, occupancy levels, cancellations
- **Network health** — Active train count per GO line at a glance
- **Push notifications** — Browser notifications when your train is delayed
- **PWA** — Install on your phone, works offline with service worker

## Architecture

Monorepo with two services:

```
Browser -> SvelteKit (SSR + proxy) -> Go API -> Metrolinx OpenData API
```

- **`api/`** — Go API server. Downloads GTFS static data at startup (refreshes every 24h), polls GTFS-RT and Metrolinx proprietary feeds in background goroutines. Merges static schedule with real-time trip updates for accurate departures.
- **`web/`** — SvelteKit frontend. Server-side rendering for initial load, client-side polling every 30s. Split-flap CSS animations for the departure board aesthetic.

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+
- A [Metrolinx OpenData API key](https://www.gotransit.com/en/open-data)

### API

```bash
cd api
export METROLINX_API_KEY=your_key_here
go run ./cmd/server/
```

Starts on port 8080.

### Web

```bash
cd web
npm install
echo "API_BASE_URL=http://localhost:8080" > .env
npm run dev
```

Starts on port 5173.

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/health` | Health check |
| GET | `/api/stops` | All GO Transit stops |
| GET | `/api/departures/{stopCode}` | Departures for a station (`?dest=` optional) |
| GET | `/api/union-departures` | Union Station departure board |
| GET | `/api/alerts` | Active service alerts |
| GET | `/api/network-health` | Active trains per GO line |
| GET | `/api/fares/{from}/{to}` | Fare info between two stations |

## Tech Stack

- **Backend:** Go stdlib (`net/http`, `slog`), GTFS protobuf parsing
- **Frontend:** SvelteKit 2, Svelte 5, Tailwind CSS 4
- **Deploy:** Railway with Railpack builder
- **CI:** GitHub Actions (Go test/vet, SvelteKit check/lint/build)

## Development

```bash
# API
cd api && go test ./... -v && go vet ./...

# Web
cd web && npm run check && npm run lint && npm run build
```
