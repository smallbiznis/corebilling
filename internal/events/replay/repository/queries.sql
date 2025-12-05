-- name: GetEventByID :one
SELECT *
FROM billing_events
WHERE id = $1
LIMIT 1;

-- name: ListEventsForTenant :many
SELECT *
FROM billing_events
WHERE tenant_id = $1
ORDER BY created_at ASC;

-- name: ListEventsByType :many
SELECT *
FROM billing_events
WHERE event_type = $1
ORDER BY created_at ASC;

-- name: ListEventsByFilters :many
SELECT *
FROM billing_events
WHERE (tenant_id = $1)
  AND (event_type = $2)
  AND (created_at >= $3)
  AND (created_at <= $4)
ORDER BY created_at ASC;
