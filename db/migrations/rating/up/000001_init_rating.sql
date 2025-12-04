CREATE TABLE IF NOT EXISTS rating_results (
    id TEXT PRIMARY KEY,
    usage_id TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_rating_usage ON rating_results (usage_id);
