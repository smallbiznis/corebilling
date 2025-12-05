CREATE TABLE IF NOT EXISTS ledger_accounts (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type SMALLINT NOT NULL,
    currency TEXT NOT NULL,
    balance_cents BIGINT NOT NULL DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ledger_accounts_tenant ON ledger_accounts (tenant_id);

CREATE TABLE IF NOT EXISTS ledger_journals (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    reference_id TEXT,
    reference_type TEXT,
    description TEXT,
    -- NOTE: idempotency handled elsewhere
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ledger_journals_tenant ON ledger_journals (tenant_id);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id TEXT PRIMARY KEY,
    journal_entry_id TEXT NOT NULL REFERENCES ledger_journals(id) ON DELETE CASCADE,
    account_id TEXT NOT NULL REFERENCES ledger_accounts(id),
    entry_type SMALLINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ledger_entries_journal ON ledger_entries (journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_account ON ledger_entries (account_id);
