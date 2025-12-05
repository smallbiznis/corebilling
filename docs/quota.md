# Quota Management

## Tenant Quota Enforcement

Tenants receive quotas defined in configuration or metadata (e.g., maximum usage events per billing period). The quota enforcement component consumes `usage.reported` events, increments counters via SQLC, and emits warning events when thresholds approach.

## Rate Limiting Model

- API gateway enforces per-tenant request limits using `tenant_id` headers.
- Internal handlers validate `usage` ingestion rate; high-water marks result in `usage.quota.warning` events.
- Dynamically adjustable limits stored in `tenant_configs` table and surfaced via Fx-provided services.

## Warning / Exceeded Events

1. `usage.quota.warning`: emitted when usage consumption exceeds 80% of quota.
2. `usage.quota.exceeded`: emitted when tenant hits quota, triggering throttling and customer notifications.
3. Handlers react by halting rating/invoice flows until limit relief is requested.

## Integration Points

- Billing API uses `tenant` service to fetch quota metadata before ingesting usage.
- Webhooks expose quota breach events to tenant dashboards.
- Scheduler can apply quota resets per billing cycle using replay module.
