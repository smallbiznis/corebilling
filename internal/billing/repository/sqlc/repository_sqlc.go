package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/billing/domain"
)

// Repository implements domain.Repository using pgx pool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a billing repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts a billing run.
func (r *Repository) Create(ctx context.Context, run domain.BillingRun) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO billing_runs (id, tenant_id, subscription_id, period_start, period_end, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		run.ID, run.TenantID, run.SubscriptionID, run.PeriodStart, run.PeriodEnd, run.Status, run.CreatedAt, run.UpdatedAt)
	return err
}

// GetByID fetches a billing run by id.
func (r *Repository) GetByID(ctx context.Context, id int64) (domain.BillingRun, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, subscription_id, period_start, period_end, status, created_at, updated_at FROM billing_runs WHERE id=$1`, id)
	var run domain.BillingRun
	if err := row.Scan(&run.ID, &run.TenantID, &run.SubscriptionID, &run.PeriodStart, &run.PeriodEnd, &run.Status, &run.CreatedAt, &run.UpdatedAt); err != nil {
		return domain.BillingRun{}, err
	}
	return run, nil
}

// ListBySubscription returns runs for a subscription.
func (r *Repository) ListBySubscription(ctx context.Context, subscriptionID int64) ([]domain.BillingRun, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, subscription_id, period_start, period_end, status, created_at, updated_at FROM billing_runs WHERE subscription_id=$1 ORDER BY period_start DESC`, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []domain.BillingRun
	for rows.Next() {
		var run domain.BillingRun
		if err := rows.Scan(&run.ID, &run.TenantID, &run.SubscriptionID, &run.PeriodStart, &run.PeriodEnd, &run.Status, &run.CreatedAt, &run.UpdatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return runs, nil
}

var _ domain.Repository = (*Repository)(nil)
