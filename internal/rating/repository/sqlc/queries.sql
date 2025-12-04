-- name: CreateRatingResult :exec
INSERT INTO rating_results (id, tenant_id, usage_id, price_id, amount_cents, currency, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,now(),now());

-- name: GetRatingResult :one
SELECT id, tenant_id, usage_id, price_id, amount_cents, currency, created_at, updated_at FROM rating_results WHERE id = $1;
