# Herald API Engine

> High-Performance Multi-Tenant Notification & Webhook Delivery Engine built with **Go 1.22+**, **Gin**, and **PostgreSQL**.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://golang.org)
[![Database](https://img.shields.io/badge/PostgreSQL-pgx-blue.svg)](https://github.com/jackc/pgx)

---

## 🏗️ 3-Repo Open Core Ecosystem

Herald is structured as a modular 3-repo ecosystem:

| Repository | Open Source? | Description & Primary Use-Case |
| :--- | :--- | :--- |
| **`Herald-API`** (This Repo) | 🟢 **Open Core (MIT)** | High-throughput Go dispatch engine, in-memory worker pool, raw SQL database layer, and tenant-scoped REST API. |
| [**`herald-js-client`**](../herald-js-client) | 🟢 **Open Source (MIT)** | Official TypeScript & JavaScript SDK (`@chi-g/herald-client`) with typed API bindings and real-time React polling hooks. |
| [**`herald-cloud`**](../herald-cloud) | 🔴 **SaaS / Commercial** | Admin Dashboard ("Dispatch Tower") for multi-tenant workspace administration, API key management, live feed analytics, and template management. |

---

## 🎯 Engine Use Cases

The `Herald-API` engine serves as centralized notification infrastructure across multiple products (such as *AuraMed*, *StoreCore*, *LMS*, *UpWearLane*, *Forafix*, and *LedgerCore*):

1. **Multi-Tenant Notification Dispatch**:
   Accepts and processes email, SMS, and push notification dispatches scoped by `tenant_id` and authorized via API keys.
2. **Asynchronous Goroutine Worker Pool**:
   In-memory buffered channels handle high-concurrency background dispatches with exponential backoff retries without external queue dependencies (no Redis/RabbitMQ required).
3. **Status Tracking & Webhook Delivery**:
   Tracks notification lifecycles (`pending` → `queued` → `sending` → `sent`/`failed`) and delivers HMAC-signed webhook callbacks on status changes.

---

## 🏛️ Architecture Rules

1. **Strict 3-Layer Pattern**: `handlers/` (HTTP binding only) → `service/` (business logic) → `repository/` (SQL only with `pgx`).
2. **Tenant Isolation**: Every database query must scope by `tenant_id`.
3. **No Global State**: Dependencies are constructed in `cmd/server/main.go` and explicitly passed down.

---

## 🛠️ Local Development & Quick Start

```bash
# 1. Start PostgreSQL with Docker
docker compose up -d postgres

# 2. Run Database Schema Migrations
psql postgresql://herald:herald@localhost:5432/herald -f schema.sql

# 3. Seed Initial Tenants & API Keys
psql postgresql://herald:herald@localhost:5432/herald -f seed.sql

# 4. Start Server
make run
# Server will start listening on http://localhost:8080
```

---

## 📄 License
MIT License &copy; Chi-G.
