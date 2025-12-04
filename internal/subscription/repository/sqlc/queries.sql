-- name: CreateSubscription :exec
INSERT INTO subscriptions (id, tenant_id, plan_id, starts_at, ends_at, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetSubscription :one
SELECT id, tenant_id, plan_id, starts_at, ends_at, status, created_at FROM subscriptions WHERE id = $1;

-- name: ListSubscriptionsByTenant :many
SELECT id, tenant_id, plan_id, starts_at, ends_at, status, created_at FROM subscriptions WHERE tenant_id = $1 ORDER BY created_at DESC;
