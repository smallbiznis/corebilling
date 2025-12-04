-- name: CreateUsageRecord :exec
INSERT INTO usage_records (id, tenant_id, subscription_id, meter_id, quantity, recorded_at, metadata, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,now(),now());

-- name: GetUsageRecord :one
SELECT id, tenant_id, subscription_id, meter_id, quantity, recorded_at, metadata, created_at, updated_at FROM usage_records WHERE id = $1;

-- name: ListUsageBySubscription :many
SELECT id, tenant_id, subscription_id, meter_id, quantity, recorded_at, metadata, created_at, updated_at FROM usage_records WHERE subscription_id = $1 ORDER BY recorded_at DESC;
