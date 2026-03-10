# Rail Six

![Services](https://github.com/teclara/railsix/actions/workflows/api.yml/badge.svg)
![Web](https://github.com/teclara/railsix/actions/workflows/web.yml/badge.svg)

Real-time GO Transit tracking for Toronto commuters. Split-flap departure board, commute dashboard with countdown timer, delay alerts, and network health — free, no ads.

## Features

- **Commute dashboard** — Save your to-work and to-home trips, see next departures with countdown timer
- **Split-flap departure board** — Full-screen board for Union Station or any GO station, scales for TV/kiosk display
- **Station lookup** — Search any GO station and view upcoming train departures
- **Real-time data** — Platform assignments, delay status, car counts, occupancy levels, cancellations
- **Network health** — Active train count per GO line at a glance
- **Delay alerts** — Browser notifications while Rail Six is open and your train becomes delayed
- **PWA** — Install on your phone, works offline with service worker

## Architecture

Monorepo with 6 microservices + shared module, connected via NATS and Redis:

```
Browser → SvelteKit (SSR) → API Gateway → departures-api/gtfs-static → Redis/Metrolinx
```

All services live under `services/`:
- **`shared/`** — Go module: models, NATS/Redis helpers, Metrolinx client
- **`gtfs-static/`** — GTFS ZIP loader, schedule queries via HTTP
- **`realtime-poller/`** — Unified poller for Metrolinx feeds → Redis + NATS
- **`departures-api/`** — Departure queries, fares, alerts, network health
- **`api-gateway/`** — Thin routing layer, CORS, health aggregation
- **`sse-push/`** — NATS → SSE streams to browsers
- **`web/`** — SvelteKit frontend

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 22+
- A [Metrolinx OpenData API key](https://www.gotransit.com/en/open-data)

### Full Stack (Docker)

```bash
docker compose up
```

### Individual Services

```bash
cd services/gtfs-static && go run .       # port 8081
cd services/realtime-poller && go run .   # polls + publishes
cd services/departures-api && go run .    # port 8082
cd services/api-gateway && go run .       # port 8080
cd services/sse-push && go run .          # port 8085
```

### Web

```bash
cd services/web
npm install
echo "API_BASE_URL=http://localhost:8080" > .env
npm run dev
```

Starts on port 5173.

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/health` | Aggregated health check |
| GET | `/api/stops` | All GO Transit stops |
| GET | `/api/departures/{stopCode}` | Departures for a station (`?dest=` optional) |
| GET | `/api/union-departures` | Union Station departure board |
| GET | `/api/alerts` | Active service alerts |
| GET | `/api/network-health` | Active trains per GO line |
| GET | `/api/fares/{from}/{to}` | Fare info between two stations |
| GET | `/api/sse` | SSE stream for real-time updates |

## Tech Stack

- **Backend:** Go stdlib (`net/http`, `slog`), NATS, Redis, GTFS protobuf parsing
- **Frontend:** SvelteKit 2, Svelte 5, Tailwind CSS 4
- **Deploy:** Railway with Railpack builder
- **CI:** GitHub Actions (Go test/vet, SvelteKit check/lint/build)

## Development

```bash
# Services
cd services && go test ./... -v -short && go vet ./...

# Web
cd services/web && npm run check && npm run lint && npm run build
```
