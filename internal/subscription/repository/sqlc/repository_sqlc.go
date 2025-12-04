package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/subscription/domain"
)

// Repository implements subscription persistence via pgxpool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts subscription.
func (r *Repository) Create(ctx context.Context, sub domain.Subscription) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO subscriptions (id, tenant_id, plan_id, starts_at, ends_at, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		sub.ID, sub.TenantID, sub.PlanID, sub.StartsAt, sub.EndsAt, sub.Status, sub.CreatedAt)
	return err
}

// GetByID fetches subscription by id.
func (r *Repository) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, plan_id, starts_at, ends_at, status, created_at FROM subscriptions WHERE id=$1`, id)
	var sub domain.Subscription
	if err := row.Scan(&sub.ID, &sub.TenantID, &sub.PlanID, &sub.StartsAt, &sub.EndsAt, &sub.Status, &sub.CreatedAt); err != nil {
		return domain.Subscription{}, err
	}
	return sub, nil
}

// ListByTenant lists subscriptions for tenant.
func (r *Repository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Subscription, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, plan_id, starts_at, ends_at, status, created_at FROM subscriptions WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.TenantID, &sub.PlanID, &sub.StartsAt, &sub.EndsAt, &sub.Status, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

var _ domain.Repository = (*Repository)(nil)
