# Event Catalog

## Event Types

| Domain | Event | Description |
| --- | --- | --- |
| Subscription | `subscription.created`, `subscription.updated`, `subscription.canceled`, `subscription.price.updated`, `subscription.status.changed` | Tracks lifecycle changes and provisioning events. |
| Usage | `usage.reported`, `usage.rated`, `usage.aggregated`, `usage.status.changed` | Meter reporting, rating completion, and aggregation readiness. |
| Rating | `rating.completed`, `rating.failed` | Finalized charge computation results. |
| Invoice | `invoice.generated`, `invoice.sent`, `invoice.paid`, `invoice.due`, `invoice.voided`, `invoice.status.changed` | Invoice lifecycle events mirrored to ledger/webhook consumers. |
| Credit & Plan | `credit.applied`, `credit.reversed`, `plan.created`, `plan.updated`, `plan.deprecated` | Metadata-level changes that impact billing behavior. |
| Scheduler | `billing.cycle.closed`, `billing.invoice.pending` | Billing cycle transitions triggered by scheduler workers. |

## Naming Conventions

- Subject format: `<resource>.<verb>[.<modifier>]` (e.g., `subscription.created`, `usage.rated`, `invoice.paid`).
- Verbs are lowercase, use dashes only when modeling weights (`usage.aggregated`).
- Version suffix (`.v1`, `.v2`) appended only when schemas need bumping; default to implicit `v1`.
- Follow-up events use `action` semantics (`invoice.generated` → `invoice.sent`).

## Event Envelope Structure

Every event is modeled by `eventv1.Event` with:

| Field | Content |
| --- | --- |
| `id` | ULID identifier generated if missing. |
| `subject` | Event subject (see table above). |
| `tenant_id` | Required, scopes data per tenant. |
| `metadata` | Struct carrying `correlation_id`, `causation_id`, `outbox_status`, and tracing context. |
| `payload` | Domain-specific struct with fields (e.g., `subscription_id`, `amount_cents`). |
| `created_at` | Timestamp of event creation. |

`events.EventEnvelope` adds helper fields: `Subject`, `TenantID`, `CorrelationID`, `CausationID`, and cached `Payload`.

## Metadata Conventions

- `correlation_id`: propagated across dispatchers and handlers to trace customer ticket. Ensured by `correlation.EnsureCorrelationID`.
- `causation_id`: set by parent event before creating follow-up events (`handler.NewFollowUpEvent`).
- `tenant_id`: always written to metadata and required by repositories, webhooks, and handlers.
- `outbox_*` keys: `outbox_status`, `outbox_retry_count`, `outbox_next_attempt_at`, `outbox_last_error` are JSON metadata entries managed by dispatcher.

## Versioning Rules

1. Increment version when you add new required payload fields or change semantics (e.g., `usage.reported.v2`).
2. Keep earlier versions active so existing consumers degrade gracefully.
3. Always update `docs/event-catalog.md` and `docs/overview.md` when adding/removing subjects.
4. Use feature flags when introducing new event chains to ensure preview deployments remain stable.
