CREATE TABLE IF NOT EXISTS rating_results (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    usage_id BIGINT NOT NULL,
    price_id BIGINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rating_usage ON rating_results (usage_id);
