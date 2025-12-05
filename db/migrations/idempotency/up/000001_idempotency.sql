CREATE TABLE IF NOT EXISTS idempotency_records (
    tenant_id       BIGINT NOT NULL,
    key             TEXT NOT NULL,
    request_hash    TEXT NOT NULL,
    response        JSONB,
    status          TEXT NOT NULL CHECK (status IN ('PROCESSING', 'COMPLETED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, key)
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_idempotency_records_updated_at
BEFORE UPDATE ON idempotency_records
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
