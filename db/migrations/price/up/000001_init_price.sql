CREATE TABLE IF NOT EXISTS prices_catalog (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    unit_amount_cents BIGINT,
    currency TEXT,
    billing_scheme SMALLINT NOT NULL DEFAULT 0,
    usage_type SMALLINT NOT NULL DEFAULT 0,
    billing_period SMALLINT NOT NULL DEFAULT 0,
    meter_id TEXT,
    tiers JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_prices_catalog_product ON prices_catalog (product_id);
CREATE INDEX IF NOT EXISTS idx_prices_catalog_meter ON prices_catalog (meter_id);
