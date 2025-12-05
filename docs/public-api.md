# Public API

## Public Endpoints

- `POST /v1/subscriptions`: Create or update subscription resources. Requires `tenant_id` header.
- `POST /v1/usage`: Ingest usage events with `idempotency_key`.
- `GET /v1/invoices`: List invoices. Supports tenant scoping.
- `POST /v1/events`: Publish custom billing events into the outbox for integrations.
- gRPC mirror services (`subscription`, `usage`, `invoice`, `webhook`) provide type-safe contracts from `third_party/go-genproto`.

## Tenant API Key Authentication

- Clients authenticate using tenant-specific API keys stored in Vault.
- Each request must include `X-API-Key` which maps to a tenant record.
- Rate limits and metrics are tagged with the tenant ID for observability.

## Rate Limiting

- Limits configured via deployment `.env` or Consul KV.
- Enforcement occurs at ingress (HTTP/gRPC).
- Violations return HTTP 429 with `Retry-After`.

## OpenAPI & SDKs

- OpenAPI spec generated from protobuf via `grpc-gateway`, available at `/openapi.yaml`.
- SDKs:
  - Go: Generated from `github.com/smallbiznis/go-genproto`.
  - Node/Python: Provided clients wrap gRPC gateways with tenant scoping helpers.
- Documentation includes sample requests/responses for each API surface.
