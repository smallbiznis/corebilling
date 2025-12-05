-- name: InsertWebhook :one
INSERT INTO webhooks (
    id, tenant_id, target_url, secret,
    event_types, enabled, created_at, updated_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8
) RETURNING id, tenant_id, target_url, secret, event_types, enabled, created_at, updated_at;

-- name: GetWebhookByID :one
SELECT id, tenant_id, target_url, secret, event_types, enabled, created_at, updated_at
FROM webhooks
WHERE id = $1;

-- name: ListWebhooksByTenant :many
SELECT id, tenant_id, target_url, secret, event_types, enabled, created_at, updated_at
FROM webhooks
WHERE tenant_id = $1
  AND enabled
  AND ($2::TEXT IS NULL OR $2 = ANY(event_types))
ORDER BY created_at DESC;

-- name: InsertDeliveryAttempt :one
INSERT INTO webhook_delivery_attempts (
    webhook_id, event_id, tenant_id, payload,
    status, attempt_no, next_run_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9
) RETURNING id, webhook_id, event_id, tenant_id, payload, status, attempt_no, next_run_at, last_error, created_at, updated_at;

-- name: UpdateDeliveryAttemptStatus :exec
UPDATE webhook_delivery_attempts
SET status = $2,
    attempt_no = $3,
    next_run_at = $4,
    last_error = $5,
    updated_at = NOW()
WHERE id = $1;

-- name: ListDueDeliveryAttempts :many
SELECT id, webhook_id, event_id, tenant_id, payload, status, attempt_no, next_run_at, last_error, created_at, updated_at
FROM webhook_delivery_attempts
WHERE status = 'PENDING'
  AND next_run_at <= $1
ORDER BY next_run_at
LIMIT $2;

-- name: MoveToDLQ :exec
WITH inserted AS (
    INSERT INTO webhook_dlq (webhook_id, event_id, tenant_id, payload, reason, created_at)
    VALUES ($1, $2, $3, $4, $5, NOW())
)
UPDATE webhook_delivery_attempts
SET status = 'DLQ',
    updated_at = NOW()
WHERE webhook_delivery_attempts.id = $6;
