CREATE TABLE IF NOT EXISTS invoice_engine_runs (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT,
    subscription_id BIGINT,
    invoice_id BIGINT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invoice_engine_runs_tenant ON invoice_engine_runs (tenant_id);
