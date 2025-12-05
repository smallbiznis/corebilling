CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    url TEXT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    secret TEXT,
    event_types TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhook_subscriptions_tenant ON webhook_subscriptions (tenant_id);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id BIGINT PRIMARY KEY,
    subscription_id BIGINT NOT NULL,
    event_id BIGINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    response_status INTEGER,
    response_body TEXT,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_subscription ON webhook_deliveries (subscription_id);
