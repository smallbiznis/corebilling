
CREATE TABLE IF NOT EXISTS billing_events (
    id              BIGSERIAL PRIMARY KEY,
    subject         TEXT NOT NULL,
    tenant_id       TEXT NOT NULL,
    resource_id     TEXT NULL,

    payload         BYTEA NOT NULL,

    status          TEXT NOT NULL DEFAULT 'pending', -- pending | dispatched | failed | dead_letter
    retry_count     INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NULL,
    last_error      TEXT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_billing_events_status
    ON billing_events(status);

CREATE INDEX IF NOT EXISTS idx_billing_events_next_attempt
    ON billing_events(next_attempt_at);
