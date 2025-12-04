CREATE TABLE IF NOT EXISTS billing_events (
    id TEXT PRIMARY KEY,
    subject TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    resource_id TEXT,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_billing_events_tenant ON billing_events (tenant_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_subject ON billing_events (subject);
