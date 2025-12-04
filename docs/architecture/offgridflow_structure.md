# OffGridFlow Repository Structure & Data Flow

## Top-level
- Entrypoints: `cmd/api` (HTTP server), `cmd/cli` (CLI utilities), `cmd/worker` (background jobs stub).
- Core domains in `internal`: config/logging/tracing, off-grid connectivity & AI routing, auth/billing, ingestion, emissions factors & calculators, compliance/reporting, API transport layers (HTTP, GraphQL), supporting utilities.
- Frontend: Next.js app under `web/app` with routes for auth, emissions explorer, compliance dashboards, workflow/setup, settings.

## Backend packages
- `internal/config`: env-driven config loader with validation and feature flags.
- `internal/logging`, `internal/tracing`: slog setup and OTEL tracing bootstrap.
- `internal/offgrid`: connectivity watcher & mode manager toggling online/offline; drives AI routing behavior.
- `internal/ai`: provider interfaces, OpenAI client, offline SimpleLocalProvider, retrying router that chooses cloud vs local based on mode.
- `internal/auth`: tenants/users/api-key models, password hashing, session/JWT manager, in-memory & Postgres stores, middleware for auth.
- `internal/billing`: Stripe client wrapper, subscription model/service, middleware enforcing subscription for protected routes.
- `internal/ingestion`: adapters for utility bills/cloud providers/CSV; activity store (Postgres or in-memory) with demo seeding; orchestration service.
- `internal/emissions`: factor registry (in-memory/Postgres), scope1/2/3 calculators, orchestrating Engine with batch support; models & tests for scopes 1/3.
- `internal/compliance`: rules/templates plus mappers/validators for CSRD, CBAM, SEC, California; maps emissions/activities into report inputs.
- `internal/reporting`: generators for Excel/PDF/XBRL outputs (structural stubs).
- `internal/api/http`: router wiring public/protected routes, handlers for auth/billing/AI chat/emissions/compliance, middleware for auth+subscription, JSON responders.
- `internal/api/graph`: GraphQL schema/resolvers (minimal/demo) for emissions/compliance.
- `internal/ai`, `internal/events`, `internal/audit`, `internal/allocation`: ancillary services (audit/event logs, allocation rules, AI types) mostly thin or stubby.
- `internal/db`: Postgres connection helper and `schema.sql` migrations.

## API data flow
1) `cmd/api` boots config → tracing → Postgres (if DSN) else in-memory stores.
2) Connectivity watcher sets `offgrid.ModeManager`; AI router picks OpenAI provider when online or offline local provider otherwise (with fallback).
3) Ingestion adapters pull activities into `ActivityStore` (seeded demo data when in-memory).
4) Emission factor registry (Postgres or in-memory) feeds scope calculators; Engine aggregates per-activity emissions.
5) HTTP router exposes:
   - Public: health, off-grid mode, auth register/login/logout, Stripe webhook.
   - Protected: auth/profile/api-key mgmt, billing checkout/status/portal, AI chat, emissions scope2 + summary, compliance CSRD + summary. Middleware enforces auth/subscription depending on config.
6) GraphQL layer exists but is secondary; not wired from main.
7) Frontend pages call REST endpoints via `web/lib/api` (e.g., `/api/emissions/scope2`).

## Frontend map (`web/app`)
- `layout.tsx`: global shell.
- Auth: `login`, `register` routes using API endpoints.
- Emissions: `emissions/page.tsx` shows Scope 2 table/summary pulled from `/api/emissions/scope2`.
- Compliance dashboards: `compliance/{csrd,cbam,sec,california}` simple views.
- Workflow/onboarding: `workflow/page.tsx` steps; Settings (`settings/*`) for billing/data sources/factors.
- Root `page.tsx`: marketing/overview entry.

## Observations & risks
- API surface leans on in-memory fallbacks; persistence, auth sessions, billing are effectively optional—production readiness hinges on DSN/Stripe/JWT envs being set.
- No CSRF/rate limiting; cookies signed but JWT secret default is weak when unset; auth middleware can be disabled via env, meaning protected routes may be reachable without checks.
- Billing/subscription middleware only activates when Stripe is configured; AI/chat and emissions endpoints otherwise unmetered.
- GraphQL package is unused in `cmd/api` wiring; reporting generators are stubs without storage/output routing.
- Compliance mappers/validators exist but handler coverage is minimal (CSRD only) and no persistence of reports.
- Tests limited to emissions/offgrid/allocation; no coverage for HTTP handlers, auth, billing, or compliance flows.
- Connectivity watcher defaults to external DNS/HTTP checks; lack of configuration could cause false offline in restricted networks.
- Frontend relies on the REST API but lacks error retries and auth token refresh; emissions UI assumes Scope 2 only and no pagination/loading states beyond basic.
