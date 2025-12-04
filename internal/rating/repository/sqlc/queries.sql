-- name: CreateRating :exec
INSERT INTO rating_results (id, usage_id, amount_cents, currency, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: ListRatingsByUsage :many
SELECT id, usage_id, amount_cents, currency, created_at FROM rating_results WHERE usage_id = $1 ORDER BY created_at DESC;
