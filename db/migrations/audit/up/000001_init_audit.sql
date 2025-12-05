CREATE TABLE IF NOT EXISTS audit_events (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    actor_type SMALLINT NOT NULL,
    actor_id BIGINT NOT NULL,
    action TEXT NOT NULL,
    action_type SMALLINT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    old_values JSONB,
    new_values JSONB,
    ip_address TEXT,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_tenant ON audit_events (tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_resource ON audit_events (resource_type, resource_id);
