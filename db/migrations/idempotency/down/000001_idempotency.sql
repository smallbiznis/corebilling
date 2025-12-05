DROP TRIGGER IF EXISTS trg_idempotency_records_updated_at ON idempotency_records;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS idempotency_records;
