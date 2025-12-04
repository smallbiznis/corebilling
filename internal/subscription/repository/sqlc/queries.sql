-- name: CreateSubscription :exec
INSERT INTO subscriptions (id, tenant_id, customer_id, price_id, status, start_at, end_at, metadata, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now(),now());

-- name: GetSubscription :one
SELECT id, tenant_id, customer_id, price_id, status, start_at, end_at, metadata, created_at, updated_at FROM subscriptions WHERE id = $1;

-- name: ListSubscriptionsByTenant :many
SELECT id, tenant_id, customer_id, price_id, status, start_at, end_at, metadata, created_at, updated_at FROM subscriptions WHERE tenant_id = $1 ORDER BY created_at DESC;
