-- name: GetRecord :one
SELECT tenant_id, key, request_hash, response, status, created_at, updated_at
FROM idempotency_records
WHERE tenant_id = $1 AND key = $2;

-- name: InsertProcessing :exec
INSERT INTO idempotency_records (
    tenant_id, key, request_hash, status
) VALUES ($1, $2, $3, 'PROCESSING');

-- name: MarkCompleted :exec
UPDATE idempotency_records
SET response = $3, status = 'COMPLETED'
WHERE tenant_id = $1 AND key = $2;
