-- name: CreateBillingRecord :exec
INSERT INTO billing_records (id, tenant_id, amount_cents, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetBillingRecord :one
SELECT id, tenant_id, amount_cents, created_at FROM billing_records WHERE id = $1;

-- name: ListBillingRecordsByTenant :many
SELECT id, tenant_id, amount_cents, created_at FROM billing_records WHERE tenant_id = $1 ORDER BY created_at DESC;
