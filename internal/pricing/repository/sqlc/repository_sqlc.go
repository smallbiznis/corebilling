package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/pricing/domain"
)

// Repository implements pricing Repository using pgxpool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a pricing repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts a pricing plan.
func (r *Repository) Create(ctx context.Context, plan domain.PricingPlan) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO pricing_plans (id, name, description, currency, amount_cents, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		plan.ID, plan.Name, plan.Description, plan.Currency, plan.AmountCents, plan.CreatedAt)
	return err
}

// GetByID retrieves a plan.
func (r *Repository) GetByID(ctx context.Context, id string) (domain.PricingPlan, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, name, description, currency, amount_cents, created_at FROM pricing_plans WHERE id=$1`, id)
	var plan domain.PricingPlan
	if err := row.Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Currency, &plan.AmountCents, &plan.CreatedAt); err != nil {
		return domain.PricingPlan{}, err
	}
	return plan, nil
}

// List returns all plans.
func (r *Repository) List(ctx context.Context) ([]domain.PricingPlan, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, description, currency, amount_cents, created_at FROM pricing_plans ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []domain.PricingPlan
	for rows.Next() {
		var p domain.PricingPlan
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Currency, &p.AmountCents, &p.CreatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return plans, nil
}

var _ domain.Repository = (*Repository)(nil)
