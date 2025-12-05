-- name: GetCycleForTenant :one
SELECT * FROM tenant_billing_cycle WHERE tenant_id = $1 LIMIT 1;

-- name: UpdateBillingCycle :exec
UPDATE tenant_billing_cycle
SET period_start = $2,
    period_end = $3,
    last_closed_at = now(),
    updated_at = now()
WHERE tenant_id = $1;

-- name: ListTenantsDueForCycleClose :many
SELECT tenant_id FROM tenant_billing_cycle
WHERE period_end <= now();
