CREATE TABLE IF NOT EXISTS invoices (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    subscription_id BIGINT NOT NULL,
    total_cents BIGINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    issued_at TIMESTAMPTZ NOT NULL,
    due_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant ON invoices (tenant_id);
