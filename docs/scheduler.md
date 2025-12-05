# Scheduler

## Billing Cycle Scheduler

The scheduler runs periodic jobs (configured via cron or timer) to evaluate `usage`/`charge` thresholds and close billing cycles per tenant. It queries the subscription database, identifies billing intervals (`monthly`, `yearly`), and emits `billing.cycle.closed` events into the Outbox.

## Cycle Computation Logic

- Determine `current_period_end` from subscriptions; if time has passed, emit a `billing.cycle.closed` event containing `subscription_id`, `tenant_id`, and `period_end`.
- Aggregates rated `usage` records to compose invoice items.
- Ensures deduplication by storing `cycle_id` metadata and using idempotent SQL updates.

## Billing Cycle Closed Event

The payload includes:

```json
{
  "cycle_id": "cycle-2025-01",
  "tenant_id": "tenant-123",
  "period_end": "2025-01-31T23:59:59Z",
  "subscription_id": "sub_abc",
  "usage_summary": {...}
}
```

Handlers listening to `billing.cycle.closed` generate invoices and ledger entries while preserving tenant isolation.

## Invoice Generation Trigger

Once the cycle closes:

1. Scheduler forwards event to invoice handler.
2. Invoice domain service composes totals, transitions state to `draft`, and emits `invoice.generated`.
3. Generated invoice flows through webhooks/ledger/outbox.
