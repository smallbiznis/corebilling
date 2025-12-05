CREATE TABLE IF NOT EXISTS usage_records (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    subscription_id BIGINT NOT NULL,
    meter_code TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    idempotency_key TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_usage_subscription ON usage_records (subscription_id);
CREATE UNIQUE INDEX IF NOT EXISTS uidx_usage_idempotency
ON usage_records (tenant_id, idempotency_key);
