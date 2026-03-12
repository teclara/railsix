# Full Application Review — Findings & Fix Plan

**Date:** 2026-03-12
**Scope:** All Go backend services + SvelteKit frontend + Railway production environment

## Context

Comprehensive code review of all Go backend services and the SvelteKit frontend. The codebase is architecturally solid — clean service boundaries, good proxy pattern, proper NATS/Redis usage. The issues below are surgical fixes, not structural problems.

Production is healthy — all 8 Railway services deployed with zero errors or warnings in logs.

---

## Critical (fix first)

### 1. `getClientIp` trusts spoofable `X-Forwarded-For` when no `ADDRESS_HEADER` configured
- **File:** `services/web/src/hooks.server.ts:12-13`
- **Risk:** Any client can forge their IP via `X-Forwarded-For` header, completely bypassing rate limiting
- **Fix:** Remove the `!env.ADDRESS_HEADER && forwardedFor` fallback branch. Only trust XFF when `ADDRESS_HEADER` is explicitly configured. Fall back to `event.getClientAddress()` otherwise.

### 2. `nsCancel()` not deferred — context leak on panic/refactor
- **File:** `services/departures-api/main.go:185-193`
- **Risk:** If code between `WithTimeout` and explicit `nsCancel()` panics, context goroutine leaks
- **Fix:** Replace `nsCancel()` at end of block with `defer nsCancel()` immediately after `context.WithTimeout`

### 3. `ServiceGlanceEntry` missing JSON struct tags
- **File:** `services/shared/models/models.go:84-93`
- **Risk:** Only model without tags. Silent breakage if fields are ever renamed; inconsistent with every other model
- **Fix:** Add `json:"tripNumber"` etc. tags matching the casing convention used by the rest of the models

---

## Important (fix before next release)

### 4. `SetHashJSON` uses non-atomic pipeline — brief empty-hash window
- **File:** `services/shared/cache/cache.go:46-59`
- **Risk:** `Pipeline()` is not transactional. Reader between `DEL` and first `HSet` sees empty hash
- **Fix:** Replace `client.Pipeline()` with `client.TxPipeline()` (wraps in `MULTI/EXEC`)

### 5. `findDelay` returns `0` for stops not in GTFS-RT update list
- **File:** `services/departures-api/departures.go:159-177`
- **Risk:** Train known to be delayed shows "On Time" for downstream stops not explicitly listed in feed
- **Fix:** Return `propagated` after the loop instead of `0`

### 6. `loadDepartures` in `MyCommute` has no abort controller
- **File:** `services/web/src/lib/components/MyCommute.svelte:64-78`
- **Risk:** Rapid direction toggling causes race — stale data from earlier direction overwrites current
- **Fix:** Add `AbortController` pattern matching `+page.svelte:80-98`

### 7. Realtime-poller health server has no graceful shutdown
- **File:** `services/realtime-poller/main.go:57-64`
- **Risk:** Health server goroutine never shut down on SIGTERM; no read/write timeouts. Inconsistent with all other services
- **Fix:** Create `http.Server` struct with timeouts, call `srv.Shutdown(shutdownCtx)` on signal

### 8. `DownloadURL` rejects exactly-50MB files
- **File:** `services/gtfs-static/store/download.go:51`
- **Risk:** `>=` check rejects valid file that is exactly 50MB
- **Fix:** Change `>=` to `>`

### 9. `Fetch` silently truncates >10MB Metrolinx responses
- **File:** `services/shared/metrolinx/client.go:44-49`
- **Risk:** Oversized response silently truncated, causing cryptic parse errors
- **Fix:** Add size check after `ReadAll` (same pattern as `DownloadURL`, with `>` not `>=`)

### 10. Inner `error()` throw swallowed by outer catch in page server loads
- **Files:** `services/web/src/routes/+page.server.ts:9-20`, `services/web/src/routes/departures/+page.server.ts:9-16`
- **Risk:** 502 error from malformed API response is re-thrown as 503 with wrong message. Inner check is dead code
- **Fix:** Re-throw SvelteKit HTTP errors in catch: `if (err instanceof Error && 'status' in err) throw err;`

### 11. `getSseUrl()` silently returns `''` in production
- **File:** `services/web/src/lib/server/proxy.ts:13-15`
- **Risk:** Inconsistent with `getBaseUrl()` which throws. Silent empty string is a latent trap
- **Fix:** Throw an error like `getBaseUrl()` does; update SSE route caller accordingly

### 12. `StationSearchInput` uses `$derived(Math.random())` for listbox ID
- **File:** `services/web/src/lib/components/StationSearchInput.svelte:21`
- **Risk:** Semantically wrong use of `$derived`; non-stable ID if ever re-evaluated; collision risk with multiple instances
- **Fix:** Use module-level counter: `let _nextId = 0;` then `const listboxId = \`listbox-${++_nextId}\`;`

---

## Minor (cleanup at convenience)

### 13. `disconnectSSE` wipes global handler arrays without notifying callers
- **File:** `services/web/src/lib/sse.ts:97-98`
- **Note:** `handlers.clear()` and `statusHandlers.length = 0` bypass individual unsub callbacks. Not currently causing bugs due to lifecycle ordering, but fragile

### 14. `GetUnionDepartureByTrip` is dead code
- **File:** `services/departures-api/redisclient.go:103-112`
- **Note:** No callers. O(n) linear scan design. Remove or leave — low risk either way

### 15. `activeRouteNames` uses `$derived` for static empty array
- **File:** `services/web/src/lib/components/MyCommute.svelte:171`
- **Note:** Should be `const activeRouteNames: string[] = []` — `$derived` on a literal is pointless

---

## Execution Order

1. **Critical #1** (XFF spoofing) — security fix, highest priority
2. **Critical #2** (defer nsCancel) — one-line fix
3. **Critical #3** (JSON tags) — add tags to ServiceGlanceEntry
4. **Important #4** (TxPipeline) — one-line change
5. **Important #5** (findDelay propagation) — one-line change
6. **Important #6** (MyCommute abort controller) — small addition
7. **Important #7-12** — remaining important fixes in any order
8. **Minor #13-15** — optional cleanup

---

## Railway Production Status (checked 2026-03-12)

**All 8 services deployed and healthy.** Zero errors or warnings in production logs.

| Service | Status | Last Deploy | Notes |
|---------|--------|-------------|-------|
| web | SUCCESS | Mar 12 03:52 | No errors in logs |
| departures-api | SUCCESS | Mar 12 01:38 | Clean INFO-only logs (stops proxy) |
| realtime-poller | SUCCESS | Mar 12 01:38 | Polling every 30s, ~140 trip updates, 17 alerts, 14-15 glance entries |
| sse-push | SUCCESS | Mar 12 01:38 | 1-3 concurrent SSE clients, clean connect/disconnect cycle |
| gtfs-static | SUCCESS | Mar 12 03:11 | GTFS loaded: 70 stops, 44 routes, 51 services |
| cache (Redis) | Running | — | — |
| message-bus (NATS) | Running | — | — |
| Monitoring | Running | — | — |

**Deployment config:** All services on Railway Pro, `us-east4-eqdc4a` region, 1 replica each. Railpack builder with Go 1.25.8. Restart policy: ON_FAILURE (max 5 retries). Health checks configured on all services (420s timeout for gtfs-static due to ZIP load).

**No production issues detected.** The code review findings above are latent bugs and hardening improvements, not active incidents.

---

## Verification

**Go services:**
```bash
cd services
go vet ./...
go test ./... -v -race -short
```

**Web:**
```bash
cd services/web
npm run test:ci && npm run check && npm run lint && npm run build
```
