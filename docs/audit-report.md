# CoreBilling BEaaS Technical Audit

## Architecture Summary
CoreBilling wires an Fx application that boots database, telemetry, event pipeline, domain modules, and both gRPC and HTTP servers. The gRPC layer exposes subscription operations, but the HTTP server only offers a readiness probe. The outbox/replay pipeline exists in code yet contains data-shape inconsistencies that prevent dispatch, and several repositories bypass tenant scoping.

## Strengths
- gRPC subscription service enforces tenant_id on incoming requests, supports pagination, and is wired into the Fx app via ModuleGRPC, showing a partial API surface with multi-tenant validation at the transport layer. 【F:internal/subscription/grpc.go†L19-L157】【F:internal/subscription/module.go†L10-L15】
- Event envelope creation hydrates correlation/causation metadata and trace context before publish, providing observability hooks and consistent tenant propagation. 【F:internal/events/envelope.go†L14-L83】
- Router wraps handlers with OTEL tracing, structured logging, bounded concurrency, and per-handler metrics recording, providing a solid execution shell for event handlers. 【F:internal/events/router/router.go†L13-L105】

## Problems Found
### [Critical] Outbox SQL shape mismatch breaks dispatch
- FetchPendingEvents selects 10 columns (status, retry_count, next_attempt_at, last_error, created_at) but scans only 6 into the struct, which will fail at runtime. The same bug exists for dead-letter reads. 【F:internal/events/outbox/repository_sqlc.go†L118-L165】【F:internal/events/outbox/repository_sqlc.go†L202-L239】
- InsertOutboxEvent writes only (id, subject, tenant_id, resource_id, payload, created_at) and never persists the status/retry columns queried later, so even if the scan bug were fixed the data model would still not satisfy dispatcher queries. 【F:internal/events/outbox/repository_sqlc.go†L111-L115】

### [Critical] HTTP/public API surface missing
- The HTTP server only exposes `/ready` and no `/v1` billing resources, leaving the documented REST surface (subscriptions, usage, invoices, events) absent. 【F:internal/server/http/server.go†L12-L55】

### [Major] Tenant isolation missing in persistence layer
- Subscription repository Get/Update operations fetch and mutate rows by id alone; tenant_id is not part of the predicate, so a tenant can read or modify another tenant’s subscription if they know the id. 【F:internal/subscription/repository/sqlc/repository_sqlc.go†L61-L101】【F:internal/subscription/repository/sqlc/repository_sqlc.go†L198-L227】

### [Major] Outbox dispatch lacks bus-integration safeguards
- Dispatcher publishes batches and marks dispatched but never validates tenant_id presence per event before publish, and failure handling moves to DLQ without recording the failing error in a separate table or emitting follow-up events. 【F:internal/events/outbox/dispatcher.go†L18-L105】

### [Minor] Missing REST gateway/OpenAPI alignment
- Although the subscription gRPC service exists, there is no grpc-gateway/OpenAPI exposure for POST /v1/subscriptions, /usage, /events, or invoice listing; the HTTP server is not bound to the gRPC surface. 【F:internal/subscription/grpc.go†L37-L157】【F:internal/server/http/server.go†L12-L55】

## Recommendations
- Fix outbox schema alignment: persist status/retry/next_attempt_at/last_error columns on insert, and scan all selected columns to avoid runtime failures. Consider deriving status from JSON metadata only or aligning the table schema/queries.
- Expose the advertised REST surface by adding grpc-gateway HTTP bindings for subscriptions, usage, invoices, and event emission, wiring them into the HTTP server alongside readiness.
- Enforce tenant_id predicates in all repositories (especially subscription Get/Update) to prevent cross-tenant reads/writes.
- Add publish-time validation in the dispatcher to reject events missing tenant_id and emit DLQ notifications or metrics when retries are exhausted.
- Wire grpc-gateway/OpenAPI generation so REST and gRPC stay consistent, and register handlers with the HTTP mux.

## Production Readiness Score (0–10)
**4 / 10** – Core telemetry and Fx wiring exist, but outbox data-shape bugs, missing REST surface, and tenant-isolation gaps block production readiness.
