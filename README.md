# CoreBilling

[![Coverage](https://codecov.io/gh/smallbiznis/corebilling/branch/main/graph/badge.svg)](https://codecov.io/gh/smallbiznis/corebilling)

CoreBilling is the billing engine for the SmallBiznis platform built with Go, Uber/Fx, and PostgreSQL. It includes SQLC-backed repositories, automatic migrations, and OpenTelemetry instrumentation.

## Features

- Clean architecture with domain/service/repository layers per billing domain.
- Uber/Fx dependency injection wiring logging, config, DB, telemetry, and servers.
- Production zap logging with global replacement.
- OpenTelemetry OTLP gRPC exporter for traces.
- Multi-service migration runner that applies migrations for enabled domains during startup.
- SQLC query definitions for each domain repository.

## Configuration

Environment variables (loaded from `.env` if present):

- `DATABASE_URL`: PostgreSQL connection string (default `postgres://postgres:postgres@localhost:5432/corebilling?sslmode=disable`).
- `SERVICE_NAME`: Service name for telemetry (default `corebilling`).
- `SERVICE_VERSION`: Version string (default `0.1.0`).
- `ENVIRONMENT`: Deployment environment (default `development`).
- `MIGRATIONS_ROOT`: Root path for migrations (default `db/migrations`).
- `ENABLED_MIGRATION_SERVICES`: Comma-separated services to migrate (default `billing,pricing,subscription,usage,rating,invoice`).
- `OTLP_ENDPOINT`: OTLP gRPC endpoint for telemetry exporter (default `localhost:4317`).

## Running migrations

Migrations are organized per service under `db/migrations/<service>/up` and `db/migrations/<service>/down`. At application startup, the migration runner scans each enabled service and executes `.sql` files in sorted order from the `up` folder.

To run migrations manually, ensure the database is reachable and start the service; migrations will be applied automatically. Place rollback scripts in the `down` folder for operational use.

## SQLC generation

SQLC configuration lives in `sqlc.yaml`. To regenerate query wrappers after modifying `queries.sql` files or schemas:

```bash
sqlc generate
```

## Running the service

1. Start a PostgreSQL database and ensure `DATABASE_URL` is set.
2. Start an OpenTelemetry Collector reachable at `OTLP_ENDPOINT` or adjust the environment variable.
3. Run the billing service:

```bash
go run ./cmd/billing
```

The service exposes:

- gRPC server on `:50051` with health check service.
- HTTP server on `:8080` with `/ready` endpoint.

## OpenTelemetry Collector

A minimal OTEL Collector configuration (not provided in repo) should accept OTLP gRPC on port `4317`. Example startup:

```bash
otelcol --config config.yaml
```

Ensure the collector exports traces to your backend of choice.
