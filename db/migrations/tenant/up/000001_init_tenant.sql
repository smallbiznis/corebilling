CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY,
    parent_id TEXT,
    type SMALLINT NOT NULL,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    status SMALLINT NOT NULL,
    default_currency TEXT,
    country_code TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug ON tenants (slug);
CREATE INDEX IF NOT EXISTS idx_tenants_parent ON tenants (parent_id);
