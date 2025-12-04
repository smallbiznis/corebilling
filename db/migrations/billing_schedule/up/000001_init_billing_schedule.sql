CREATE TABLE IF NOT EXISTS billing_schedules (
    id BIGSERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    subscription_id TEXT NOT NULL,
    period TEXT NOT NULL,
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_billing_schedules_subscription ON billing_schedules (subscription_id);
