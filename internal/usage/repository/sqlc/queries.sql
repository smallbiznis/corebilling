-- name: CreateUsage :exec
INSERT INTO usage_records (id, subscription_id, metric, quantity, recorded_at)
VALUES ($1, $2, $3, $4, $5);

-- name: ListUsageBySubscription :many
SELECT id, subscription_id, metric, quantity, recorded_at FROM usage_records WHERE subscription_id = $1 ORDER BY recorded_at DESC;
