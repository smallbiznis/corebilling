-- name: CreateInvoice :exec
INSERT INTO invoices (id, tenant_id, subscription_id, total_cents, status, issued_at, due_at, metadata, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now(),now());

-- name: GetInvoice :one
SELECT id, tenant_id, subscription_id, total_cents, status, issued_at, due_at, metadata, created_at, updated_at FROM invoices WHERE id = $1;

-- name: ListInvoicesByTenant :many
SELECT id, tenant_id, subscription_id, total_cents, status, issued_at, due_at, metadata, created_at, updated_at FROM invoices WHERE tenant_id = $1 ORDER BY issued_at DESC;
