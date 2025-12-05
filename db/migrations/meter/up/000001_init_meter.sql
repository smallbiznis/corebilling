CREATE TABLE IF NOT EXISTS meters (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    code TEXT NOT NULL,
    aggregation SMALLINT NOT NULL DEFAULT 0,
    transform SMALLINT NOT NULL DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_meters_tenant_code ON meters (tenant_id, code);
