CREATE TABLE IF NOT EXISTS customers (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    external_reference TEXT,
    email TEXT,
    name TEXT,
    phone TEXT,
    currency TEXT,
    billing_address JSONB,
    shipping_address JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_customers_tenant ON customers (tenant_id);
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_tenant_external ON customers (tenant_id, external_reference);
