package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/billing/domain"
)

// Repository implements domain.Repository using pgx and SQLC-style queries.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a billing repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts a billing record.
func (r *Repository) Create(ctx context.Context, record domain.BillingRecord) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO billing_records (id, tenant_id, amount_cents, created_at) VALUES ($1,$2,$3,$4)`,
		record.ID, record.TenantID, record.AmountCents, record.CreatedAt)
	return err
}

// GetByID fetches a billing record by id.
func (r *Repository) GetByID(ctx context.Context, id string) (domain.BillingRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, amount_cents, created_at FROM billing_records WHERE id=$1`, id)
	var rec domain.BillingRecord
	if err := row.Scan(&rec.ID, &rec.TenantID, &rec.AmountCents, &rec.CreatedAt); err != nil {
		return domain.BillingRecord{}, err
	}
	return rec, nil
}

// ListByTenant returns records belonging to a tenant.
func (r *Repository) ListByTenant(ctx context.Context, tenantID string) ([]domain.BillingRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, amount_cents, created_at FROM billing_records WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.BillingRecord
	for rows.Next() {
		var rec domain.BillingRecord
		if err := rows.Scan(&rec.ID, &rec.TenantID, &rec.AmountCents, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

var _ domain.Repository = (*Repository)(nil)
