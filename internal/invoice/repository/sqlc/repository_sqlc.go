package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/invoice/domain"
)

// Repository handles invoice persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts invoice.
func (r *Repository) Create(ctx context.Context, inv domain.Invoice) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, billing_period_start, billing_period_end, total_cents, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		inv.ID, inv.TenantID, inv.BillingPeriodStart, inv.BillingPeriodEnd, inv.TotalCents, inv.Status, inv.CreatedAt)
	return err
}

// GetByID fetches invoice.
func (r *Repository) GetByID(ctx context.Context, id string) (domain.Invoice, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, billing_period_start, billing_period_end, total_cents, status, created_at FROM invoices WHERE id=$1`, id)
	var inv domain.Invoice
	if err := row.Scan(&inv.ID, &inv.TenantID, &inv.BillingPeriodStart, &inv.BillingPeriodEnd, &inv.TotalCents, &inv.Status, &inv.CreatedAt); err != nil {
		return domain.Invoice{}, err
	}
	return inv, nil
}

// ListByTenant lists invoices for tenant.
func (r *Repository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Invoice, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, billing_period_start, billing_period_end, total_cents, status, created_at FROM invoices WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		if err := rows.Scan(&inv.ID, &inv.TenantID, &inv.BillingPeriodStart, &inv.BillingPeriodEnd, &inv.TotalCents, &inv.Status, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return invoices, nil
}

var _ domain.Repository = (*Repository)(nil)
