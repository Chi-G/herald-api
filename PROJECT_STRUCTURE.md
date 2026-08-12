# Herald — Project Structure & Architecture

Multi-tenant notification & webhook delivery API. Go + Gin + PostgreSQL + goroutine worker pool.
Serves any client identically: React admin dashboard, Flutter app, or server-to-server (AuraMed, StoreCore, LMS, UpWearLane, Forafix, LedgerCore).


## Folder layout

```
herald/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point — wires config, DB, router, worker pool, starts HTTP server
│
├── internal/                       # Private application code (not importable by other modules)
│   ├── config/
│   │   └── config.go                # Loads env vars into a typed Config struct
│   │
│   ├── database/
│   │   ├── postgres.go              # DB connection pool (pgx or database/sql)
│   │   └── migrations/
│   │       ├── 0001_init.up.sql
│   │       └── 0001_init.down.sql
│   │
│   ├── models/
│   │   ├── tenant.go
│   │   ├── api_key.go
│   │   ├── notification.go
│   │   ├── notification_attempt.go
│   │   └── webhook.go
│   │
│   ├── repository/                  # DB access layer — one file per aggregate, raw SQL via pgx/sqlx
│   │   ├── tenant_repo.go
│   │   ├── notification_repo.go
│   │   └── webhook_repo.go
│   │
│   ├── service/                     # Business logic — orchestrates repos + worker pool
│   │   ├── notification_service.go
│   │   ├── webhook_service.go
│   │   └── auth_service.go
│   │
│   ├── handlers/                    # Gin HTTP handlers — thin, just bind/validate/call service
│   │   ├── notification_handler.go
│   │   ├── webhook_handler.go
│   │   ├── tenant_handler.go
│   │   └── health_handler.go
│   │
│   ├── middleware/
│   │   ├── auth.go                  # API key + JWT verification
│   │   ├── rate_limit.go            # Per-tenant token bucket
│   │   ├── logger.go                # Structured request logging
│   │   └── recovery.go              # Panic recovery -> JSON error
│   │
│   ├── worker/
│   │   ├── pool.go                  # Goroutine worker pool (N workers, buffered job channel)
│   │   ├── dispatcher.go            # Picks provider based on channel, handles retry/backoff
│   │   └── providers/
│   │       ├── email.go             # SMTP / Resend / SES adapter
│   │       ├── sms.go                # Termii / Twilio adapter (Termii for NG numbers)
│   │       └── push.go               # FCM adapter (for the Flutter app side)
│   │
│   └── router/
│       └── router.go                 # Route groups, middleware wiring
│
├── pkg/                              # Code safe to share/reuse across projects
│   ├── jwtutil/
│   │   └── jwt.go
│   └── apierror/
│       └── errors.go                 # Standard JSON error envelope
│
├── .env.example
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml                # herald + postgres, for local dev
├── Makefile                          # make run / make migrate / make test
├── README.md
├── .cursorrules                      # Cursor AI coding rules
└── GEMINI.md                         # Gemini CLI / Code Assist rules
```

## Why this layout

- **`cmd/` vs `internal/`** — standard Go project layout. `cmd/server/main.go` stays tiny (wiring only); all real logic lives in `internal/`, which Go's compiler enforces can't be imported by outside modules.
- **Repository → Service → Handler** — same three-layer separation you already use in Laravel (Model/Repository → Service → Controller) and FastAPI (CRUD → Service → Route). Nothing new conceptually, just Go syntax.
- **`worker/` is the heart of the project** — this is where goroutines, channels, and `sync` actually get used for something real, not a toy example.
- **`pkg/` vs `internal/`** — `pkg/jwtutil` and `pkg/apierror` are generic enough you could lift them into your *next* Go project (e.g. AuraMed's Go microservice, if you ever build one).

## Request flow (typical)

```
Client (React/Flutter)
   -> POST /api/v1/notifications   (API key auth via middleware)
   -> handlers/notification_handler.go   (binds + validates JSON)
   -> service/notification_service.go    (creates DB record, status=pending)
   -> worker/pool.go                     (job pushed to channel)
   -> worker goroutine picks it up
   -> worker/dispatcher.go               (retry/backoff logic)
   -> providers/email.go|sms.go|push.go  (actually sends)
   -> repository updates status -> sent/failed
   -> webhook fired to tenant's configured URL (if set)
```

Client polls `GET /api/v1/notifications/:id` or listens on the webhook — no external queue (Redis/RabbitMQ) needed at this scale, since Go's goroutines + channels do that job in-process.

## Tech choices

| Concern | Choice | Why |
|---|---|---|
| Framework | Gin | Fast to ship, huge middleware ecosystem — matches your Laravel/Express instincts |
| DB | PostgreSQL | You already run Postgres (Neon) elsewhere — consistent with JournalGrid |
| DB access | `pgx` (or `sqlx` if you want closer-to-SQL control) | Skip GORM here deliberately — this project is partly to learn writing real SQL in Go, GORM would hide it |
| Auth | API key (tenant-to-tenant) + JWT (admin dashboard users) | Two different trust levels, two different mechanisms |
| Migrations | `golang-migrate` | Simple, CLI-driven, works with plain `.sql` files |
| Config | env vars via `.env` (loaded with `godotenv` in dev) | Matches your Laravel `.env` muscle memory |
