-- name: Create :exec
INSERT INTO meters (id, tenant_id, code, aggregation, transform, metadata, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);

-- name: GetByID :one
SELECT id, tenant_id, code, aggregation, transform, metadata, created_at, updated_at FROM meters WHERE id=$1;

-- name: ListByTenant :many
SELECT id, tenant_id, code, aggregation, transform, metadata, created_at, updated_at FROM meters WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: Update :exec
UPDATE meters SET tenant_id=$2, code=$3, aggregation=$4, transform=$5, metadata=$6, updated_at=$7 WHERE id=$1;