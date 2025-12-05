CREATE TABLE IF NOT EXISTS domains (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    domain TEXT NOT NULL,
    verification_method SMALLINT NOT NULL,
    verification_status SMALLINT NOT NULL,
    verification_token TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    ssl_status SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_domains_tenant_domain ON domains (tenant_id, domain);
