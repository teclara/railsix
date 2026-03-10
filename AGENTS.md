# Repository Guidelines

## Project Structure & Module Organization
`services/` contains all deployable services. Go microservices live in `services/departures-api/`, `services/gtfs-static/`, `services/realtime-poller/`, `services/sse-push/`, with shared Go code in `services/shared/`. The Go workspace file is `services/go.work`. The SvelteKit web app proxies all API calls server-side — there is no public-facing API gateway.

`services/web/` contains the SvelteKit app. Routes live in `services/web/src/routes`, shared UI in `services/web/src/lib/components`, helpers in `services/web/src/lib`, and static assets in `services/web/static`. Repo-level docs live in `docs/` and the top-level `README.md`.

## Build, Test, and Development Commands
Backend:

```bash
cd services
go vet ./...
go test ./... -v -short
```

Frontend:

```bash
cd services/web
npm install
npm run dev
npm run check
npm run lint
npm run build
```

Use `METROLINX_API_KEY` for the API and `API_BASE_URL=http://localhost:8080` for the web app.

## Coding Style & Naming Conventions
Go code should stay `gofmt`-clean and package-focused; prefer small internal packages over cross-package leakage. Svelte/TypeScript formatting is enforced by Prettier and ESLint in `services/web/`. Prettier uses tabs, single quotes, no trailing commas, and `printWidth: 100`.

Use PascalCase for Svelte components (`SplitFlapBoard.svelte`), SvelteKit route conventions (`+page.svelte`, `+page.server.ts`), and descriptive lower-case Go package names.

## Testing Guidelines
Backend tests use Go's `testing` package. Add or update tests for every behavior change; prefer table-driven tests where they clarify edge cases.

The frontend currently relies on `npm run check`, `npm run lint`, and `npm run build` in CI rather than a dedicated test runner. Treat those as the minimum gate for web changes.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit style with optional scopes, for example `fix(web): ...`, `style(web): ...`, and `feat: ...`. Keep commits focused by service when possible.

PRs should include a short summary, affected area, any env/config changes, and screenshots for visible UI changes. Before opening a PR, run the same checks GitHub Actions runs for the service you touched.
