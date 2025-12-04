-- name: CreatePricingPlan :exec
INSERT INTO pricing_plans (id, name, description, currency, amount_cents, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetPricingPlan :one
SELECT id, name, description, currency, amount_cents, created_at FROM pricing_plans WHERE id = $1;

-- name: ListPricingPlans :many
SELECT id, name, description, currency, amount_cents, created_at FROM pricing_plans ORDER BY created_at DESC;
