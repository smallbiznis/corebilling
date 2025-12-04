-- name: CreateUsageRecord :exec
INSERT INTO usage_records (
    id, tenant_id, customer_id, subscription_id,
    meter_code, value, recorded_at, idempotency_key,
    metadata, created_at, updated_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, now(), now()
);

-- name: ListUsageBySubscription :many
SELECT
    id, tenant_id, customer_id, subscription_id,
    meter_code, value, recorded_at, idempotency_key,
    metadata, created_at, updated_at
FROM usage_records
WHERE subscription_id = $1
ORDER BY recorded_at DESC
LIMIT $2 OFFSET $3;
