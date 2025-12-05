# Onboarding

## Getting Started

1. Clone repo and install Go 1.25.x plus sqlc (`go install github.com/kyleconroy/sqlc/cmd/sqlc@latest`).
2. Copy `.env.example` to `.env` and set `DATABASE_URL`, `KAFKA_BROKERS`/`NATS_URL`, `OTLP_ENDPOINT`.
3. Run `sqlc generate` after editing `.sql` files.
4. Launch PostgreSQL (via Docker/Compose) and run `go run cmd/billing/main.go`.

## Project Structure

- `internal/app`: Fx application wiring.
- `internal/*/domain`: Business logic and state machines.
- `internal/*/repository/sqlc`: SQLC-generated persistence layer.
- `internal/events`: Outbox, bus, router, handler abstractions.
- `db/migrations/<service>`: Service-specific SQL migrations.
- `third_party/go-genproto`: Generated protobuf clients (do not edit directly).

## Running Locally

Use `docker-compose up postgres kafka nats` (if required), then:

```bash
go run cmd/billing/main.go
```

Set `ENABLE_MIGRATIONS=true` via environment scripts to auto-run migrations.

## Testing Event Flow End-to-End

1. Call `usage` gRPC to insert records.
2. Verify `billing_events` entries via `psql`.
3. Watch dispatcher logs for `event.publish`.
4. Confirm handlers (e.g., invoice) execute by checking `invoices` table and emitted webhooks.

## Common Pitfalls

- Forgetting to add new migrations under `db/migrations/<service>`.
- Editing generated protos instead of upstream (`third_party/go-genproto`).
- Missing `tenant_id` in events, which fails authorization and repository checks.
