# Webhooks

## Webhook Subscription Model

Each tenant manages webhooks via `internal/webhook/domain`. A webhook entry contains:

- `id`: GUID primary key.
- `tenant_id`: required to isolate delivery attempts.
- `target_url`: HTTP endpoint.
- `secret`: HMAC secret for signing.
- `event_types`: `TEXT[]` filter for subjects.
- `enabled`: toggle for delivery.

The JSON contract is defined in SQLC and bound to `webhook.Config`, enabling per-tenant delivery limits via `WEBHOOK_WORKER_INTERVAL_SECONDS` and `WEBHOOK_WORKER_LIMIT`.

## Security Model (HMAC)

Payloads are signed using `sbwh_sig=v1:<hex>` where `<hex>` is the hexadecimal HMAC-SHA256 digest of the serialized event payload, computed with the tenant’s `secret`. The worker sends this signature in the `X-Signature` header along with `X-Event-ID`, `X-Event-Type`, and `X-Tenant-ID`.

## Delivery Workflow

1. `webhook.Service.DispatchForEvent` filters webhooks by `tenant_id` and subject, enqueues `webhook_delivery_attempts`.
2. Worker (`webhook.Worker`) polls attempts, builds HTTP POST, attaches headers/signature, and sends payload.
3. Successful responses mark status `SUCCESS` and increment metrics (`corebilling_webhook_delivery_success_total`).
4. Failures increment attempt count and update `next_run_at`.

## Retry & DLQ

- Base delay 10 seconds (`WEBHOOK_BASE_DELAY_SECONDS`) with exponential growth (factor 2) and jitter capped by `WEBHOOK_MAX_DELAY_SECONDS`.
- After `WEBHOOK_MAX_RETRIES` attempts, the attempt moves to `webhook_dlq` and status `DLQ`.
- Metrics track failures and DLQ hits via Prometheus counters.

## Tenant-level Configuration

Each tenant registers webhooks through gRPC (`webhook.Service`) or a future HTTP portal. Secrets rotate by updating the `secret` column, and event filters support `ANY(event_types)`.

## Example Payload

```json
{
  "id": "evt_abc",
  "subject": "invoice.generated",
  "tenant_id": "tenant_123",
  "payload": {
    "invoice_id": "inv_001",
    "total_cents": 50000,
    "currency": "USD"
  },
  "metadata": {
    "correlation_id": "corr_123",
    "causation_id": "evt_usage"
  }
}
```

Signed headers:

- `X-Signature: sbwh_sig=v1:<hex>`
- `X-Tenant-ID: tenant_123`
- `X-Event-ID: evt_abc`
- `X-Event-Type: invoice.generated`
