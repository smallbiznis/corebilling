CREATE TABLE IF NOT EXISTS invoices (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    customer_id TEXT,
    subscription_id TEXT,
    status SMALLINT NOT NULL DEFAULT 0,
    currency_code TEXT NOT NULL DEFAULT 'USD',
    total_cents BIGINT NOT NULL DEFAULT 0,
    subtotal_cents BIGINT NOT NULL DEFAULT 0,
    tax_cents BIGINT NOT NULL DEFAULT 0,
    invoice_number TEXT,
    issued_at TIMESTAMPTZ,
    due_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant ON invoices (tenant_id);
CREATE INDEX IF NOT EXISTS idx_invoices_customer ON invoices (customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_subscription ON invoices (subscription_id);
