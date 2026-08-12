-- Herald database schema
-- PostgreSQL 15+
-- Run via golang-migrate, or directly: psql -d herald -f schema.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- for gen_random_uuid()

-- =========================================================
-- TENANTS — each product you plug in (AuraMed, StoreCore, LMS...)
-- =========================================================
CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =========================================================
-- API KEYS — how a tenant authenticates server-to-server calls
-- =========================================================
CREATE TABLE api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,           -- e.g. "AuraMed Production"
    key_hash     VARCHAR(255) NOT NULL UNIQUE,     -- store SHA-256 hash, never the raw key
    key_prefix   VARCHAR(12) NOT NULL,             -- first N chars shown in dashboard for identification, e.g. "hrld_live_ab"
    is_active    BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at   TIMESTAMPTZ
);
CREATE INDEX idx_api_keys_tenant ON api_keys(tenant_id);

-- =========================================================
-- ADMIN USERS — humans who log into the dashboard (JWT auth)
-- =========================================================
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50) NOT NULL DEFAULT 'admin',  -- admin | viewer
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =========================================================
-- NOTIFICATIONS — the core entity
-- =========================================================
CREATE TYPE notification_channel AS ENUM ('email', 'sms', 'push');
CREATE TYPE notification_status  AS ENUM ('pending', 'queued', 'sending', 'sent', 'failed', 'retrying', 'cancelled');
CREATE TYPE notification_priority AS ENUM ('low', 'normal', 'high', 'urgent');

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel         notification_channel NOT NULL,
    priority        notification_priority NOT NULL DEFAULT 'normal',
    status          notification_status NOT NULL DEFAULT 'pending',

    recipient       VARCHAR(255) NOT NULL,      -- email address, phone number, or device/FCM token
    subject         VARCHAR(500),                -- used for email; null for sms/push
    body            TEXT NOT NULL,
    metadata        JSONB NOT NULL DEFAULT '{}', -- arbitrary tenant data, e.g. {"order_id": "123"}

    scheduled_at    TIMESTAMPTZ,                 -- null = send immediately
    sent_at         TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,

    attempt_count   INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 5,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_tenant       ON notifications(tenant_id);
CREATE INDEX idx_notifications_status       ON notifications(status);
CREATE INDEX idx_notifications_scheduled    ON notifications(scheduled_at) WHERE status = 'pending';
CREATE INDEX idx_notifications_created      ON notifications(created_at);

-- =========================================================
-- NOTIFICATION ATTEMPTS — audit trail of every send try (retry/backoff history)
-- =========================================================
CREATE TABLE notification_attempts (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id    UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    attempt_number     INT NOT NULL,
    status             VARCHAR(20) NOT NULL,       -- success | failed
    provider           VARCHAR(50),                 -- e.g. "resend", "termii", "fcm"
    provider_response  JSONB,
    error_message      TEXT,
    attempted_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_attempts_notification ON notification_attempts(notification_id);

-- =========================================================
-- WEBHOOKS — where tenants get notified of delivery events
-- =========================================================
CREATE TABLE webhooks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    secret      VARCHAR(255) NOT NULL,          -- used to HMAC-sign payloads, tenant verifies authenticity
    events      TEXT[] NOT NULL DEFAULT '{}',   -- e.g. {'notification.sent','notification.failed'}
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_webhooks_tenant ON webhooks(tenant_id);

-- =========================================================
-- WEBHOOK DELIVERIES — audit trail of outbound webhook calls
-- =========================================================
CREATE TABLE webhook_deliveries (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id       UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    notification_id  UUID REFERENCES notifications(id) ON DELETE SET NULL,
    event            VARCHAR(100) NOT NULL,
    payload          JSONB NOT NULL,
    response_code    INT,
    response_body    TEXT,
    succeeded        BOOLEAN NOT NULL DEFAULT false,
    attempted_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);

-- =========================================================
-- RATE LIMIT CONFIG — per-tenant throughput limits
-- =========================================================
CREATE TABLE rate_limit_configs (
    tenant_id           UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    requests_per_minute INT NOT NULL DEFAULT 60,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =========================================================
-- Trigger: auto-update updated_at columns
-- =========================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
