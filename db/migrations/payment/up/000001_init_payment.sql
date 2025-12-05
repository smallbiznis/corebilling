CREATE TABLE IF NOT EXISTS payment_methods (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    provider SMALLINT NOT NULL,
    type SMALLINT NOT NULL,
    display_name TEXT,
    last4 TEXT,
    exp_month TEXT,
    exp_year TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    provider_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payment_methods_tenant ON payment_methods (tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_provider ON payment_methods (provider);

CREATE TABLE IF NOT EXISTS payment_attempts (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    invoice_id BIGINT NOT NULL,
    payment_method_id BIGINT NOT NULL,
    status SMALLINT NOT NULL,
    provider_transaction_id TEXT,
    attempted_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payment_attempts_invoice ON payment_attempts (invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_attempts_status ON payment_attempts (status);
