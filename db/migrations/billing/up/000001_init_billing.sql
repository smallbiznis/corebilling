CREATE TABLE IF NOT EXISTS billing_records (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_billing_records_tenant ON billing_records (tenant_id);
