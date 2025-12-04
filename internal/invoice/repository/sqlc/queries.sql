-- name: CreateInvoice :exec
INSERT INTO invoices (id, tenant_id, billing_period_start, billing_period_end, total_cents, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetInvoice :one
SELECT id, tenant_id, billing_period_start, billing_period_end, total_cents, status, created_at FROM invoices WHERE id = $1;

-- name: ListInvoicesByTenant :many
SELECT id, tenant_id, billing_period_start, billing_period_end, total_cents, status, created_at FROM invoices WHERE tenant_id = $1 ORDER BY created_at DESC;
