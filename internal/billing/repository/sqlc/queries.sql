-- name: InsertBillingRun :exec
INSERT INTO billing_runs (id, tenant_id, subscription_id, period_start, period_end, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now(), now());

-- name: GetBillingRun :one
SELECT id, tenant_id, subscription_id, period_start, period_end, status, created_at, updated_at
FROM billing_runs WHERE id = $1;

-- name: ListBillingRunsBySubscription :many
SELECT id, tenant_id, subscription_id, period_start, period_end, status, created_at, updated_at
FROM billing_runs WHERE subscription_id = $1 ORDER BY period_start DESC;
