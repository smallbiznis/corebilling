package sqlc

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/meter/domain"
)

// Repository manages meter persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a meter repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, meter domain.Meter) error {
	metadata, err := marshalJSON(meter.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `INSERT INTO meters (id, tenant_id, code, aggregation, transform, metadata, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		meter.ID,
		meter.TenantID,
		meter.Code,
		meter.Aggregation,
		meter.Transform,
		metadata,
		meter.CreatedAt,
		meter.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Meter, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, code, aggregation, transform, metadata, created_at, updated_at FROM meters WHERE id=$1`, id)
	var meter domain.Meter
	var metadata []byte
	if err := row.Scan(&meter.ID, &meter.TenantID, &meter.Code, &meter.Aggregation, &meter.Transform, &metadata, &meter.CreatedAt, &meter.UpdatedAt); err != nil {
		return domain.Meter{}, err
	}
	meter.Metadata = jsonToMap(metadata)
	return meter, nil
}

func (r *Repository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]domain.Meter, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, code, aggregation, transform, metadata, created_at, updated_at FROM meters WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meters []domain.Meter
	for rows.Next() {
		var meter domain.Meter
		var metadata []byte
		if err := rows.Scan(&meter.ID, &meter.TenantID, &meter.Code, &meter.Aggregation, &meter.Transform, &metadata, &meter.CreatedAt, &meter.UpdatedAt); err != nil {
			return nil, err
		}
		meter.Metadata = jsonToMap(metadata)
		meters = append(meters, meter)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return meters, nil
}

func (r *Repository) Update(ctx context.Context, meter domain.Meter) error {
	metadata, err := marshalJSON(meter.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `UPDATE meters SET tenant_id=$2, code=$3, aggregation=$4, transform=$5, metadata=$6, updated_at=$7 WHERE id=$1`,
		meter.ID,
		meter.TenantID,
		meter.Code,
		meter.Aggregation,
		meter.Transform,
		metadata,
		meter.UpdatedAt,
	)
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

var _ domain.Repository = (*Repository)(nil)
