CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    subject TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    data JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_tenant ON events (tenant_id);
CREATE INDEX IF NOT EXISTS idx_events_subject ON events (subject);
