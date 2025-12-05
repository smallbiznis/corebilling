CREATE TABLE IF NOT EXISTS tax_rules (
    id BIGINT PRIMARY KEY,
    region_code TEXT NOT NULL,
    name TEXT NOT NULL,
    rate_percent DOUBLE PRECISION NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tax_rules_region ON tax_rules (region_code);
