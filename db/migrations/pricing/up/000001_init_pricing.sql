CREATE TABLE IF NOT EXISTS products (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_tenant_code ON products (tenant_id, code);

CREATE TABLE IF NOT EXISTS prices (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL REFERENCES products(id),
    code TEXT NOT NULL,
    lookup_key TEXT,
    pricing_model SMALLINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL,
    unit_amount_cents BIGINT NOT NULL,
    billing_interval SMALLINT NOT NULL DEFAULT 0,
    billing_interval_count INTEGER NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_prices_product ON prices (product_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_prices_tenant_code ON prices (tenant_id, code);
