# Replay Engine

## Replay Rules

- Only events with `outbox_status` of `dispatched` or `dead_letter` are replayable.
- Replay preserves original `correlation_id` and `tenant_id`.
- `replay.ReplayService` (internal) loads events, optionally updates metadata, and re-inserts them into the outbox with `pending` status.

## Safety Constraints

- Idempotency trackers ensure repeated replays do not mutate subscriber state multiple times within the tracking window.
- Handlers rely on state machines to reject invalid transitions even when replayed events arrive.
- Replay scripts must never alter `billing_events` payload schema; to upgrade, create a new event version.

## Replay Flow

1. Retrieve archived event from `billing_events` or DLQ.
2. Validate tenant ownership and the presence of a valid API key.
3. Use replay service to re-enqueue the event; OpenTelemetry traces log replay reason.
4. Dispatcher picks up event again and publishes to the router, which reruns handlers.

## Using the Replay Module

Use CLI automation (e.g., `./scripts/replay --tenant=... --id=... --reason="fix billing"`). The module will log per-tenant metrics, update `last_replayed_at`, and emit `billing.replay.finished`.
