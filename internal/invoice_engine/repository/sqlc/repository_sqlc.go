package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/invoice_engine/domain"
)

// Repository persists invoice engine runs.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs an invoice engine run repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, run domain.Run) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO invoice_engine_runs (id, tenant_id, customer_id, subscription_id, invoice_id, period_start, period_end, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		run.ID,
		run.TenantID,
		nullIfEmpty(run.CustomerID),
		nullIfEmpty(run.SubscriptionID),
		run.InvoiceID,
		run.PeriodStart,
		run.PeriodEnd,
		run.CreatedAt,
	)
	return err
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

var _ domain.Repository = (*Repository)(nil)
