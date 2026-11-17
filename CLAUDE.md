# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Nurtura is a child health-tracking app (growth metrics, vaccines, activity logs) with three components in one monorepo:

- `server/` — Go API (the active backend)
- `web/` — Svelte 5 + TypeScript + Vite frontend (the active client)
- `mobile/` — Flutter project scaffold, currently just boilerplate (`mobile/lib/main.dart`), not yet built out

## Dev environment

The repo uses a Nix flake for the dev shell (`flake.nix`), providing Go, Node, Bun, PostgreSQL, Flutter, Android SDK, and language servers. Entering the shell runs `nix-hook.fish`, which initializes a local Postgres cluster under `.pgdata/` (socket at `.pgsock/`, user `postgres`, db `postgres`, trust auth) and starts it if not already running, then opens a tmux session.

```bash
nix develop            # enter the dev shell (sets up + starts local Postgres)
psql                   # connects to the local instance via PGHOST/PGUSER/PGDATABASE env vars set by nix-hook.fish
```

Load the schema into a fresh local DB with `psql -f server/queries/schema.sql` (and optionally the seed files in `server/queries/seeds/`).

## Server (`server/`, Go)

```bash
cd server
go run ./cmd/server          # run locally; loads server/.env via godotenv, defaults to :8080
go test ./tests/...          # integration tests against a running server (see below)
go build ./...
```

Required env vars: `DATABASE_URL`, `JWT_SECRET` (must be ≥32 chars). Optional: `PORT`, `CORS_ORIGINS` (comma-separated, used only by the Vercel handler).

**Two entrypoints, one `app.New()`:**
- `cmd/server/main.go` — plain `net/http` server for local dev (default `:8080`). **This is the API to develop against.**
- `api/index.go` — Vercel serverless function handler (`vercel.json` rewrites all paths to `/api`).

Both wrap the handler with `app.WithCORS` (`app/cors.go`, `rs/cors`), which allows `http://localhost:5173` plus any comma-separated `CORS_ORIGINS`.

**Don't use `vercel dev` as the local API.** It spawns a separate Go process per build and routes concurrent requests to processes that may not be listening yet, so parallel requests return sporadic **502**s (the dashboard fires four at once via `Promise.all` and reliably trips this). The same load against `go run ./cmd/server` is clean. Point `web/.env` at `http://localhost:8080` and use `cmd/server`; reserve `vercel dev` for checking the serverless path itself.

Both call `app.New()` in `app/app.go`, which builds the pgx pool and route table. Routes are plain Go 1.22+ `http.ServeMux` patterns (`"POST /patients/{id}/records"`), wrapped with `a.requireAuth` for anything needing a logged-in user.

**Request flow / conventions** (see `app/utils.go`, `app/auth.go`, `app/patients.go` as references):
- `decode()` enforces a 1MB body cap and `DisallowUnknownFields`.
- Path UUIDs are validated via `pathUUID()`/`uuidRe` before hitting the DB.
- Ownership is enforced in the SQL itself, not in app logic — e.g. `loadPatient` filters `WHERE id = $1 AND user_id = $2`; there is no separate authorization layer.
- `serverErr()` logs the real error server-side and always returns a generic 500 to the client; never leak internal errors or request bodies to responses/logs.
- `userId(r)` reads the JWT subject out of request context (set by `requireAuth`); auth uses `golang-jwt/jwt/v5` HS256 tokens, subject = user UUID, checked against `uuidRe`.
- Login/register deliberately equalizes timing for unknown vs. wrong-password cases using a precomputed `dummyHash` — preserve this if touching `auth.go`.
- `pagination()` in `utils.go` is the shared helper for `limit`/`offset` query params (limit 1–200, default 50).

**Schema** (`server/queries/schema.sql`): single `BEGIN`/`COMMIT` migration defining `users`, `patients`, `metrics`, `records`, `logs`, `vaccines`, `patient_vaccines`, `growth_reference`, and a `patient_activity` table kept current via triggers (`touch_patient_child`/`touch_patient_self`) on insert/update/delete of child rows. Row-level security is enabled on every table (policies aren't in this file — check for them separately before assuming a query is unrestricted by RLS). There's no migration tool wired up yet; this file is the whole schema.

**Tests** (`server/tests/api_test.go`): black-box HTTP integration tests, not unit tests. They hit a live server at `BASE_URL` (default `http://localhost:3000`) — start the server first. One shared user/token (`sharedToken`) is reused across tests that don't need isolation, since `/auth/login` is rate-limited (~10/min); `login()` auto-retries once after a 61s sleep on a 429. Keep using `sharedToken`/`newUser` helpers rather than registering fresh users per test unless isolation is actually needed.

## Web (`web/`, Svelte 5 + TS + Vite)

```bash
cd web
bun install
bun run dev             # vite dev server
bun run build
bun run check            # svelte-check + tsc (no-emit type checking)
bun run lint             # prettier --check + eslint
bun run format           # prettier --write
```

Requires `VITE_API_URL` in `web/.env` pointing at the Go server.

- No SvelteKit/router — `App.svelte` just toggles between the login page and main page based on whether `getToken()` (from `routes/login/auth.ts`) returns a valid, non-expired token stored in `localStorage`.
- `$src` / `$src/*` is a TS path alias for `src/` (see `tsconfig.app.json`); use it instead of deep relative imports.
- `routes/login/auth.ts` is the only API client code so far: `post()` wraps `fetch`, throws on non-2xx using the server's `{"error": "..."}` shape, `authFetch()` attaches the bearer token and auto-logs-out on 401.
- Formatting: tabs, single quotes, no trailing commas, 100 print width (`prettier.config.js`) — run `bun run format` rather than hand-formatting.
