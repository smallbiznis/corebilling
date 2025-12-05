CREATE TABLE IF NOT EXISTS webhooks (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    target_url TEXT NOT NULL,
    secret TEXT NOT NULL,
    event_types TEXT[] NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_tenant ON webhooks (tenant_id);

CREATE TABLE IF NOT EXISTS webhook_delivery_attempts (
    id BIGINT PRIMARY KEY,
    webhook_id BIGINT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL,
    attempt_no INT NOT NULL,
    next_run_at TIMESTAMPTZ NOT NULL,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS webhook_dlq (
    id BIGINT PRIMARY KEY,
    webhook_id BIGINT NOT NULL,
    event_id BIGINT NOT NULL,
    tenant_id TEXT NOT NULL,
    payload JSONB NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhook_delivery_attempts_status
    ON webhook_delivery_attempts(status);

CREATE INDEX IF NOT EXISTS idx_webhook_delivery_attempts_next_run
    ON webhook_delivery_attempts(next_run_at);
