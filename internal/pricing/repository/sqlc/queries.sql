CREATE SCHEMA IF NOT EXISTS pricing;

CREATE TABLE IF NOT EXISTS pricing.products (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    description TEXT NULL,
    active BOOLEAN NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS pricing.prices (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL REFERENCES pricing.products(id),
    code TEXT NOT NULL,
    lookup_key TEXT NULL,
    pricing_model SMALLINT NOT NULL,
    currency TEXT NOT NULL,
    unit_amount_cents BIGINT NOT NULL,
    billing_interval SMALLINT NOT NULL,
    billing_interval_count INT NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS pricing.price_tiers (
    id BIGINT PRIMARY KEY,
    price_id BIGINT NOT NULL REFERENCES pricing.prices(id),
    start_quantity DOUBLE PRECISION NOT NULL,
    end_quantity DOUBLE PRECISION NOT NULL,
    unit_amount_cents BIGINT NOT NULL,
    unit TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- name: InsertProduct :one
INSERT INTO pricing.products (
    id,
    tenant_id,
    name,
    code,
    description,
    active,
    metadata
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING *;

-- name: GetProductByID :one
SELECT *
FROM pricing.products
WHERE tenant_id = $1 AND id = $2;

-- name: ListProductsByTenant :many
SELECT *
FROM pricing.products
WHERE tenant_id = $1
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: InsertPrice :one
INSERT INTO pricing.prices (
    id,
    tenant_id,
    product_id,
    code,
    lookup_key,
    pricing_model,
    currency,
    unit_amount_cents,
    billing_interval,
    billing_interval_count,
    active,
    metadata
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
) RETURNING *;

-- name: InsertPriceTier :exec
INSERT INTO pricing.price_tiers (
    id,
    price_id,
    start_quantity,
    end_quantity,
    unit_amount_cents,
    unit,
    metadata
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
);

-- name: GetPriceByID :one
SELECT *
FROM pricing.prices
WHERE tenant_id = $1 AND id = $2;

-- name: ListPricesWithFilter :many
SELECT *
FROM pricing.prices
WHERE tenant_id = $1
  AND ($2::BIGINT IS NULL OR product_id = $2)
  AND ($3::TEXT IS NULL OR code = $3)
ORDER BY id
LIMIT $4 OFFSET $5;

-- name: ListPriceTiersByPriceIDs :many
SELECT *
FROM pricing.price_tiers
WHERE price_id = ANY($1::BIGINT[]);
