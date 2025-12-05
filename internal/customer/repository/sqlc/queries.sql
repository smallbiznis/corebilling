-- name: Create :exec
INSERT INTO customers (id, tenant_id, external_reference, email, name, phone, currency, billing_address, shipping_address, metadata, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);

-- name: GetByID :one
SELECT id, tenant_id, external_reference, email, name, phone, currency, billing_address, shipping_address, metadata, created_at, updated_at FROM customers WHERE id=$1;

-- name: ListByTenant :many
SELECT id, tenant_id, external_reference, email, name, phone, currency, billing_address, shipping_address, metadata, created_at, updated_at FROM customers WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: Update :exec
UPDATE customers SET tenant_id=$2, external_reference=$3, email=$4, name=$5, phone=$6, currency=$7, billing_address=$8, shipping_address=$9, metadata=$10, updated_at=$11 WHERE id=$1;