-- name: Create :exec
INSERT INTO tenants (
			id, parent_id, type, name, slug, status, default_currency, country_code, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);

-- name: GetByID :one
SELECT id, parent_id, type, name, slug, status, default_currency, country_code, metadata, created_at, updated_at
		FROM tenants WHERE id=$1;

-- name: List :many
SELECT id, parent_id, type, name, slug, status, default_currency, country_code, metadata, created_at, updated_at
		FROM tenants WHERE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;