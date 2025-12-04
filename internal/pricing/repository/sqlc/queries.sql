-- name: InsertProduct :one
INSERT INTO products (
    id, tenant_id, name, code, description, active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetProductByID :one
SELECT *
FROM products
WHERE tenant_id = $1 AND id = $2;

-- name: ListProductsByTenant :many
SELECT *
FROM products
WHERE tenant_id = $1
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: InsertPrice :one
INSERT INTO prices (
    id, tenant_id, product_id, code,
    lookup_key, pricing_model, currency,
    unit_amount_cents, billing_interval,
    billing_interval_count, active, metadata
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9,
    $10, $11, $12
) RETURNING *;

-- name: GetPriceByID :one
SELECT *
FROM prices
WHERE tenant_id = $1 AND id = $2;

-- name: ListPricesWithFilter :many
SELECT *
FROM prices
WHERE tenant_id = $1
  AND ($2::BIGINT IS NULL OR product_id = $2)
  AND ($3::TEXT IS NULL OR code = $3)
ORDER BY id
LIMIT $4 OFFSET $5;
