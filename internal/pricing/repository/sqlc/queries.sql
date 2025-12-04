-- name: CreateProduct :exec
INSERT INTO products (id, tenant_id, name, code, description, active, metadata, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now());

-- name: GetProduct :one
SELECT id, tenant_id, name, code, description, active, metadata, created_at, updated_at FROM products WHERE id = $1;

-- name: ListProductsByTenant :many
SELECT id, tenant_id, name, code, description, active, metadata, created_at, updated_at FROM products WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: CreatePrice :exec
INSERT INTO prices (id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now(),now());

-- name: GetPrice :one
SELECT id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at FROM prices WHERE id = $1;

-- name: ListPricesByProduct :many
SELECT id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at FROM prices WHERE product_id = $1 ORDER BY created_at DESC;
