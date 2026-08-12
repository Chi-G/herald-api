# Herald

A multi-tenant notification & webhook delivery API built in Go + Gin, serving as shared infrastructure across Forahia Solutions' public and private applications instead of relying on third-party SaaS like SendGrid or Novu.

> 🚀 **Project Status**: Upcoming release (Active Development). The admin dashboard and frontend integration client are currently being built using **React** (`react-client/`).

### Connected Tenant Applications
- **AuraMed** — [auramed.cc](https://auramed.cc)
- **StoreCore** — *(Not live yet)*
- **Forahia LMS** — [lms.forahia.com](https://lms.forahia.com)
- **UpWearLane** — [UpWearLane.com](https://UpWearLane.com)
- **Forafix** — [forahia.com/forafix](https://forahia.com/forafix)
- **LedgerCore** — [ledger.forahia.com](https://ledger.forahia.com)



## What's in this scaffold

- `PROJECT_STRUCTURE.md` — full folder layout and architecture rationale
- `schema.sql` — complete PostgreSQL schema (tenants, api_keys, notifications,
  notification_attempts, webhooks, webhook_deliveries, rate_limit_configs)
- `.cursorrules` — rules for Cursor AI to follow this project's conventions
- `GEMINI.md` — equivalent rules for Gemini CLI / Code Assist
- `internal/` — the actual Go source, partially scaffolded:
  - `models/notification.go` — the core struct + request DTO
  - `worker/pool.go` + `worker/dispatcher.go` — the concurrency engine (goroutines + channels)
  - `middleware/auth.go` — API key authentication
  - `handlers/notification_handler.go` — HTTP layer
  - `router/router.go` — route wiring
  - `pkg/apierror/errors.go` — standard JSON response envelope
- `react-client/` — a typed TypeScript client + React hook + example component
  ready to drop into any of your admin dashboards

## What's NOT scaffolded yet (build these next, in order)

1. `internal/config/config.go` — env var loading
2. `internal/database/postgres.go` — pgx connection pool setup
3. `internal/repository/*.go` — the actual SQL queries (notification_repo.go is
   referenced by dispatcher.go and the handler — build this next, it's the
   natural next step and a good place to practice raw SQL in Go)
4. `internal/service/notification_service.go` — orchestrates repo + worker pool
5. `internal/worker/providers/email.go` — start with a simple SMTP or Resend adapter
6. `cmd/server/main.go` — wire everything together and start the server

## Suggested build order

Build bottom-up so each piece is testable in isolation:
1. `models` -> `repository` (get DB reads/writes working, test with `go run` + psql)
2. `worker` providers (get one channel, e.g. email, actually sending)
3. `service` (glue repo + worker together)
4. `handlers` + `router` + `middleware` (expose it over HTTP)
5. `cmd/server/main.go` (wire it all up)
6. React client — point `NotificationSender.example.tsx` at your running server

## Local dev

```bash
# Start Postgres
docker compose up -d postgres

# Run migrations
psql postgresql://herald:herald@localhost:5432/herald -f schema.sql

# Run the server (once main.go exists)
go run cmd/server/main.go
```

## Next milestone after v1

Once notifications work end-to-end (create -> dispatch -> status update), the
natural v2 features are: webhook delivery on status change (HMAC-signed payloads),
a scheduled-notification cron sweep (`scheduled_at` field is already in the schema),
and a simple admin dashboard in React showing delivery stats per tenant.
