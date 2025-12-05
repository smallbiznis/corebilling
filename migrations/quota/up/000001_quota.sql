CREATE TABLE IF NOT EXISTS tenant_quota_limits (
    tenant_id TEXT PRIMARY KEY,
    max_events_per_day BIGINT NOT NULL,
    max_usage_units BIGINT NOT NULL,
    soft_warning_threshold FLOAT DEFAULT 0.8,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tenant_quota_usage (
    tenant_id TEXT PRIMARY KEY,
    events_today BIGINT NOT NULL DEFAULT 0,
    usage_units BIGINT NOT NULL DEFAULT 0,
    reset_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION reset_tenant_quota_usage()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.reset_at IS NULL OR NEW.reset_at::date < now()::date THEN
        NEW.events_today := 0;
        NEW.usage_units := 0;
        NEW.reset_at := now();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_reset_tenant_quota_usage ON tenant_quota_usage;
CREATE TRIGGER trg_reset_tenant_quota_usage
BEFORE INSERT OR UPDATE ON tenant_quota_usage
FOR EACH ROW
EXECUTE FUNCTION reset_tenant_quota_usage();
