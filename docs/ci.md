# Continuous Integration for BEaaS

This repository uses a multi-stage CI pipeline focused on reliability and fast feedback. All workflows live under `.github/workflows/` and are orchestrated by `ci.yaml`.

## Workflow topology

`ci.yaml` orchestrates the following stages:

1. **Lint** (`lint.yaml`) – formatting, linters, vet, and `sqlc` generation checks.
2. **Unit tests** (`test-unit.yaml`) – fast correctness checks with coverage.
3. **Race detector** (`test-race.yaml`) – `go test -race` for data races.
4. **Integration tests** (`test-integration.yaml`) – runs `internal/idempotency` tests against Testcontainers-powered Postgres/Redis/NATS/Kafka.
5. **End-to-end tests** (`test-e2e.yaml`) – boots the BEaaS stack with Docker Compose and runs `tests/e2e`.
6. **SQL drift detection** (`sql-drift.yaml`) – ensures `sqlc` outputs match migrations and configuration.
7. **Coverage aggregation** – merges coverage from unit/integration suites and reports to Codecov.
8. **Docker build & publish** (`build.yaml` using `reusable-build.yaml`) – multi-arch buildx to GHCR.
9. **Preview deployment (optional)** – gated job that can deploy ephemeral environments when enabled.
10. **Security scanning** (`security.yaml`) – Trivy (filesystem + image) and Snyk (Go modules).

## Local integration testing

1. Ensure Docker is running and supports user namespace cloning (`sudo sysctl -w kernel.unprivileged_userns_clone=1`).
2. Run `scripts/setup-testcontainers.sh` to export the required environment for Testcontainers.
3. Start dependencies if needed or let Testcontainers manage them.
4. Execute the integration suite:

```bash
POSTGRES_VERSION=16 REDIS_VERSION=7 \
  go test ./internal/idempotency/... -tags=integration -coverprofile=int.out
```

The reusable workflow mirrors this flow by starting containers for Postgres, Redis, NATS, and Kafka.

## Testcontainers requirements

- Docker available in the environment (local or CI runner).
- `kernel.unprivileged_userns_clone=1` (handled automatically in CI).
- Network access to mapped ports: Postgres (5432), Redis (6379), NATS (4222), Kafka (9092).
- Optional overrides via `TESTCONTAINERS_HOST_OVERRIDE` and `TESTCONTAINERS_RYUK_DISABLED`.

## Interpreting CI results

- **Lint failures**: Fix formatting (`gofmt`, `gofumpt`) or lint issues; regenerate `sqlc` outputs if stale.
- **Race detector**: Investigate data races; rerun locally with `go test -race ./...`.
- **Integration/E2E**: Check service readiness and dependency connectivity; the reusable Testcontainers workflow publishes DSNs as job outputs.
- **SQL drift**: Run `scripts/sql-diff.sh` locally; reconcile migrations with generated code.
- **Coverage**: Inspect merged coverage uploaded to Codecov; individual suites publish artifacts.
- **Build**: Review buildx logs and resulting GHCR tags (`ghcr.io/<org>/corebilling:<sha>` and `:latest`).
- **Security**: Trivy and Snyk reports highlight vulnerabilities; update dependencies or base images accordingly.

## SQL drift detection

`sql-drift.yaml` runs `scripts/sql-diff.sh`, which regenerates `sqlc` outputs and fails if any diffs against `db/`, `internal/`, or `sqlc.yaml` are detected. Resolve by updating migrations or committing regenerated code.

## Preview deployments

The CI orchestrator exposes an optional preview job. Enable it by either:

- Setting the `deploy_preview` input on `workflow_dispatch`, or
- Adding a `preview` label to a pull request.

The job uses the built image from the CI build stage and can be extended to point at your preview environment of choice (e.g., ephemeral Kubernetes namespace or a temporary VM).
