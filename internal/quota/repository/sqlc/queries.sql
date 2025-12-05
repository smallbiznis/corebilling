-- name: GetQuotaLimit :one
SELECT * FROM tenant_quota_limits WHERE tenant_id = $1 LIMIT 1;

-- name: GetQuotaUsage :one
SELECT * FROM tenant_quota_usage WHERE tenant_id = $1 LIMIT 1;

-- name: UpsertQuotaUsage :exec
INSERT INTO tenant_quota_usage (tenant_id, events_today, usage_units, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (tenant_id)
DO UPDATE SET events_today = $2, usage_units = $3, updated_at = now();

-- name: ListTenantsOverLimit :many
SELECT tenant_quota_usage.tenant_id
FROM tenant_quota_usage
JOIN tenant_quota_limits
  ON tenant_quota_limits.tenant_id = tenant_quota_usage.tenant_id
WHERE events_today > max_events_per_day
   OR usage_units > max_usage_units;
