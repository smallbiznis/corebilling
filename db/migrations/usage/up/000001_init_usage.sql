CREATE TABLE IF NOT EXISTS usage_records (
    id TEXT PRIMARY KEY,
    subscription_id TEXT NOT NULL,
    metric TEXT NOT NULL,
    quantity BIGINT NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_usage_subscription ON usage_records (subscription_id);
