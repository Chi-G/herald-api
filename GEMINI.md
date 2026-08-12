# GEMINI.md — Herald project context

## What this project is
Herald is a multi-tenant notification & webhook delivery API written in Go, using the
Gin web framework and PostgreSQL. It sends email/SMS/push notifications on behalf of
multiple tenant products (AuraMed, StoreCore, LMS, UpWearLane, Forafix, LedgerCore) and delivers webhook

callbacks when a notification's status changes.

## Who you're helping
An experienced backend engineer (Laravel, FastAPI, ~6 years) actively learning Go.
Explain idiomatic Go patterns briefly when they differ meaningfully from PHP/Python
conventions (goroutines vs async/await, channels vs promises, explicit error returns
vs exceptions, interfaces vs abstract classes). Keep explanations short — a sentence
or two, not a lecture — unless asked to go deeper.

## Architecture rules to respect
1. Three layers: `handlers/` (HTTP binding only) -> `service/` (business logic) ->
   `repository/` (SQL only). Do not collapse these layers "for simplicity."
2. Raw SQL via `pgx`, no ORM. This is intentional — do not suggest GORM.
3. Every notification-related query must scope by `tenant_id`. Treat a missing
   tenant_id filter as a security bug, not a nitpick.
4. Concurrency for dispatch goes through the worker pool in `internal/worker/`
   (goroutines + buffered channels). Do not introduce Redis, RabbitMQ, or other
   external queues — that defeats the purpose of this project.
5. All HTTP responses use the shared error/success envelope in `pkg/apierror`.

## Coding standards
- Go 1.22+, stdlib-first, gofmt-clean.
- Explicit error wrapping with `%w`, never swallowed errors.
- `context.Context` propagated through every I/O call.
- No package-level global state — dependencies constructed in `cmd/server/main.go`
  and passed down explicitly.
- Table-driven tests for service/repository logic.

## When generating code
- Prefer complete, runnable file contents over fragments when scaffolding a new file.
- When modifying an existing file, show only the diff/relevant section unless the
  whole file changed significantly.
- Flag any new third-party dependency explicitly before adding it to go.mod.
- If a request would violate the layering rules above (e.g. "just query the DB
  directly in the handler for speed"), point that out rather than complying silently.
