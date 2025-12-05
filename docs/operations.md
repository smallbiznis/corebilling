# Operations

## Scaling Recommendations

- **Dispatcher Pool:** Run multiple dispatcher instances with shared PostgreSQL access; `billing_events` partitioned by tenant_id and indexing on `status` helps shards select only pending rows.
- **Router Workers:** Configure Kafka/NATS consumer groups per region; align `events.Router` groups (`cfg.KafkaGroupID`) with handler throughput.
- **Webhook Workers:** Horizontal workers read `webhook_delivery_attempts`, commit `next_run_at`, and back pressure webhooks by tuning `WEBHOOK_WORKER_LIMIT`.

## Worker Pool Sizes

- Starting point: 2 dispatchers, 3 routers, 1 webhook worker per tenant region.
- Adjust per `corebilling_outbox_dispatcher_inflight` and webhook latency metrics.

## Dispatcher Sharding

- Shard by tenant prefix (e.g., `tenantA-` vs `tenantB-`) via bus subjects.
- Align `NATSStream` subjects per tenant for multi-tenant isolation.

## DLQ Operations

- Inspect `webhook_dlq` for permanent webhook failures. Use stored `reason` and `payload` to troubleshoot.
- Use `billing_events` DLQ to replay via `replay.ReplayService`.
- Prometheus alerts on `corebilling_webhook_delivery_dlq_total`.

## Replay Workflow

1. Query `billing_events` for status `dead_letter`.
2. Validate replay reason with compliance.
3. Run replay CLI with `--tenant-id` to re-enqueue event.

## Monitoring Dashboards

- Prometheus metrics to monitor:
  - `corebilling_outbox_dispatcher_dispatch_count`
  - `corebilling_webhook_delivery_success_total`
  - `corebilling_webhook_delivery_failure_total`
  - `corebilling_webhook_delivery_dlq_total`
  - `corebilling_event_router_handle_duration`
  - `corebilling_usage_ingestion_rate`

## Logging & Tracing

- Structured logs include `tenant_id`, `event_id`, `webhook_id`, `subscription_id`.
- Traces propagated via `OpenTelemetry` (span names prefixed with handler subject).
- Use `ctxlogger` to attach `event.subject`.
