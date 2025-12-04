CREATE TABLE IF NOT EXISTS credit_notes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    invoice_id TEXT,
    type SMALLINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL,
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_credit_notes_tenant ON credit_notes (tenant_id);
CREATE INDEX IF NOT EXISTS idx_credit_notes_invoice ON credit_notes (invoice_id);
