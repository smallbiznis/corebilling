package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	billingcyclerepo "github.com/smallbiznis/corebilling/internal/billingcycle/repository/sqlc"
)

// SQLRepository implements Repository using sqlc-generated queries.
type SQLRepository struct {
	queries *billingcyclerepo.Queries
}

// NewRepository constructs a repository backed by sqlc queries.
func NewRepository(pool *pgxpool.Pool) Repository {
	db := stdlib.OpenDB(*pool.Config().ConnConfig)
	return &SQLRepository{queries: billingcyclerepo.New(db)}
}

func (r *SQLRepository) GetCycleForTenant(ctx context.Context, tenantID string) (BillingCycle, error) {
	rec, err := r.queries.GetCycleForTenant(ctx, tenantID)
	if err != nil {
		return BillingCycle{}, err
	}
	var lastClosed *time.Time
	if rec.LastClosedAt.Valid {
		t := rec.LastClosedAt.Time
		lastClosed = &t
	}
	return BillingCycle{
		TenantID:     rec.TenantID,
		PeriodStart:  rec.PeriodStart,
		PeriodEnd:    rec.PeriodEnd,
		LastClosedAt: lastClosed,
	}, nil
}

func (r *SQLRepository) UpdateBillingCycle(ctx context.Context, params UpdateBillingCycleParams) error {
	return r.queries.UpdateBillingCycle(ctx, billingcyclerepo.UpdateBillingCycleParams{
		TenantID:    params.TenantID,
		PeriodStart: params.PeriodStart,
		PeriodEnd:   params.PeriodEnd,
	})
}

func (r *SQLRepository) ListTenantsDueForCycleClose(ctx context.Context) ([]string, error) {
	rows, err := r.queries.ListTenantsDueForCycleClose(ctx)
	if err != nil {
		return nil, err
	}
	tenants := make([]string, 0, len(rows))
	tenants = append(tenants, rows...)
	return tenants, nil
}

var _ Repository = (*SQLRepository)(nil)
