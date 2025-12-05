package sqlc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	tenantv1 "github.com/smallbiznis/go-genproto/smallbiznis/tenant/v1"

	"github.com/smallbiznis/corebilling/internal/tenant/domain"
)

// Repository stores tenants.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a tenant repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, tenant domain.Tenant) error {
	metadata, err := marshalJSON(tenant.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO tenants (
			id, parent_id, type, name, slug, status, default_currency, country_code, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, tenant.ID, nullInt64(tenant.ParentID), int16(tenant.Type), tenant.Name, tenant.Slug, int16(tenant.Status), nullString(tenant.DefaultCurrency), nullString(tenant.CountryCode), metadata, tenant.CreatedAt, tenant.UpdatedAt)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Tenant, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, parent_id, type, name, slug, status, default_currency, country_code, metadata, created_at, updated_at
		FROM tenants WHERE id=$1
	`, id)

	var tenant domain.Tenant
	var parentID sql.NullInt64
	var metadata []byte
	var typ, status int16
	if err := row.Scan(
		&tenant.ID,
		&parentID,
		&typ,
		&tenant.Name,
		&tenant.Slug,
		&status,
		&tenant.DefaultCurrency,
		&tenant.CountryCode,
		&metadata,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	); err != nil {
		return domain.Tenant{}, err
	}
	if parentID.Valid {
		tenant.ParentID = parentID.Int64
	}
	tenant.Type = tenantv1.TenantType(typ)
	tenant.Status = tenantv1.TenantStatus(status)
	tenant.Metadata = jsonToMap(metadata)
	return tenant, nil
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]domain.Tenant, error) {
	args := make([]interface{}, 0, 6)
	conds := make([]string, 0, 2)
	idx := 1

	if filter.ParentID != "" {
		parsedParentID, err := strconv.ParseInt(filter.ParentID, 10, 64)
		if err != nil {
			return nil, err
		}
		conds = append(conds, fmt.Sprintf("parent_id=$%d", idx))
		args = append(args, parsedParentID)
		idx++
	}
	if len(conds) == 0 {
		conds = append(conds, "TRUE")
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 25
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	limitIdx := idx
	args = append(args, limit)
	idx++
	offsetIdx := idx
	args = append(args, offset)
	idx++

	query := fmt.Sprintf(`
		SELECT id, parent_id, type, name, slug, status, default_currency, country_code, metadata, created_at, updated_at
		FROM tenants WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, strings.Join(conds, " AND "), limitIdx, offsetIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var tenant domain.Tenant
		var parentID sql.NullInt64
		var metadata []byte
		var typ, status int16
		if err := rows.Scan(
			&tenant.ID,
			&parentID,
			&typ,
			&tenant.Name,
			&tenant.Slug,
			&status,
			&tenant.DefaultCurrency,
			&tenant.CountryCode,
			&metadata,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if parentID.Valid {
			tenant.ParentID = parentID.Int64
		}
		tenant.Type = tenantv1.TenantType(typ)
		tenant.Status = tenantv1.TenantStatus(status)
		tenant.Metadata = jsonToMap(metadata)
		tenants = append(tenants, tenant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *Repository) Update(ctx context.Context, tenant domain.Tenant) error {
	metadata, err := marshalJSON(tenant.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE tenants SET parent_id=$2, type=$3, name=$4, slug=$5, status=$6,
			default_currency=$7, country_code=$8, metadata=$9, updated_at=$10
		WHERE id=$1
	`, tenant.ID, nullInt64(tenant.ParentID), int16(tenant.Type), tenant.Name, tenant.Slug, int16(tenant.Status),
		nullString(tenant.DefaultCurrency), nullString(tenant.CountryCode), metadata, tenant.UpdatedAt)
	return err
}

func marshalJSON(value map[string]interface{}) ([]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return json.Marshal(value)
}

func jsonToMap(value []byte) map[string]interface{} {
	if len(value) == 0 {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(value, &data); err != nil {
		return nil
	}
	return data
}

func nullString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nullInt64(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

var _ domain.Repository = (*Repository)(nil)
