# Outbox Pattern

## How Outbox Works

1. Domain services insert events via `outbox.ApplyMetadata` using SQLC repositories.
2. Events persist in `billing_events` with metadata-driven state.
3. Dispatcher polls pending rows, serializes them into `events.EventEnvelope`, publishes through the configured bus (Kafka/NATS/Noop), and updates metadata (`outbox_status`, `outbox_retry_count`).

## DB Schema

`billing_events` stores:

| Column | Purpose |
| --- | --- |
| `id` | Event ULID, primary key. |
| `subject` | Event subject to route. |
| `tenant_id` | Multi-tenant scoping column; non-null. |
| `resource_id` | Optional domain key (subscription_id, invoice_id). |
| `payload` | Serialized JSON of `eventv1.Event`. |
| `created_at` | Timestamp for audit. |

Metadata keys injected into the JSON payload:

- `outbox_status`: `pending`, `dispatched`, `failed`, `dead_letter`.  
- `outbox_retry_count`: how many retries have attempted.  
- `outbox_next_attempt_at`: when next retry should run.  
- `outbox_last_error`: last failure reason.

## Dispatcher Lifecycle

1. `outbox.Dispatcher.Run` wakes every 2 seconds (configurable) and calls `processBatch`.
2. `FetchPendingEvents` returns rows whose `outbox_status = pending` and `next_attempt_at <= now`.
3. Successful publish updates metadata to `dispatched`.
4. Failures increment `outbox_retry_count` and compute `next_attempt_at` via exponential backoff.
5. DLQ transition occurs when retry limit reached (`outbox.ShouldMoveToDLQ`).

## Retry/Backoff/DLQ Flow

- Base delay: 1s, max delay: 5 minutes.
- Retry count increments by 1 per failure; `ComputeNextAttempt` caps jitter/delay.
- On `ShouldMoveToDLQ`, dispatcher calls `MoveToDeadLetter` storing final error.
- Dispatcher logs structured errors with `event_id`, `tenant_id`, and OpenTelemetry traces to facilitate replays.

## Idempotency Guarantees

- Handlers use `outbox.IdempotencyTracker` to reject duplicates within a configurable window (default 10 minutes).
- Each event carries `correlation_id`/`causation_id`, so downstream Webhooks and replay flows can deduplicate repeated delivery attempts.
- State machines guard transitions to prevent repeated `subscription.created` → `subscription.trialing` loops.

## Operational Notes

- Monitor Prometheus metrics `corebilling_outbox_dispatcher_inflight`, `corebilling_outbox_dispatcher_errors`, and `corebilling_webhook_delivery_*`.
- Ensure `ENABLED_MIGRATION_SERVICES` includes `billing_event` to create `billing_events`.
- For replay, re-insert or update payload metadata with `outbox_status=pending`, resetting `outbox_retry_count`.
