# Overview

## What is BEaaS?

BEaaS (Event-Driven Billing-as-a-Service) is a multi-tenant, cloud-native billing platform built on Go, Uber/Fx, SQLC, and PostgreSQL. It models billing interactions as events so downstream subsystems (rate engines, metering, invoicing, webhooks) stay resilient under percentile traffic spikes, retries, and cross-tenant isolation requirements.

The platform ingests usage telemetry, rating adjustments, and subscription lifecycle changes as canonical events, persists them via an Outbox pattern, routes them through clean architecture handlers, and emits follow-up events for invoicing, ledger entries, and external consumers. NATS and Kafka serve as pluggable transport backbones depending on deployment, while Kafka/NATS consumers are instrumented with OpenTelemetry and structured logging.

## High-level Architecture

BEaaS centers on four pillars:

1. **Event Ingestion & Outbox:** Each domain (subscription, usage, rating, invoice) writes events into `billing_events` using SQLC-generated repositories, guaranteeing transactional durability and multi-tenant scoping by requiring `tenant_id` on every payload.
2. **Dispatcher & Router:** An `outbox.Dispatcher` reads pending events and publishes them through a bus interface (`events.Bus`). The `events.Router` subscribes handlers per subject, ensuring idempotent processing with correlation/causation tracing.
3. **Domain Services & State Machines:** Services (subscription, usage, invoice) enforce state transitions, apply business rules, and emit new events into the outbox when needed. State machines validate lifecycle flows such as `subscription.created -> subscription.trialing -> subscription.active`.
4. **Delivery & Observability:** Webhooks, scheduler workers, and webhook delivery engines consume derived events. Prometheus metrics, tracing through OpenTelemetry, and structured logging provide production-grade observability.

## Why Event-Driven Billing?

- **Resilience:** Events decouple producers and consumers, allowing downstream consumers (e.g., invoice generator, ledger writer) to retry asynchronously without blocking the producer.
- **Auditability:** Every action is encoded as a versioned event with metadata (`correlation_id`, `causation_id`, `tenant_id`), making tracing and troubleshooting deterministic.
- **Scalability:** Event batching + Outbox allows horizontal scaling by sharding dispatchers and router handlers across regions.
- **Extensibility:** New capabilities (plans, credits, external webhooks) plug into the pipeline by registering new handlers and webhooks against existing subjects.

## Key Capabilities

- **Multi-tenant semantics** enforced per event, webhook, and database namespace.
- **SQLC-backed persistence** ensures type-safe access to migrations for every domain (`db/migrations/<service>`).
- **Uber/Fx modules** wire services, dispatchers, routers, webhook workers, and telemetry in a predictable lifecycle.
- **OpenTelemetry + Prometheus integration** for tracing, metrics, and logging across event boundaries.
- **Pluggable transport layer** (NATS, Kafka, Noop) isolating event publishers from domain logic.
- **Webhook engine with HMAC signing** delivers events to tenant-provided endpoints with exponential backoff and DLQ.
- **Scheduler & replay infrastructure** ensure billing cycles close deterministically and failures can be reprocessed safely.
