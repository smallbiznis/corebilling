package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	quotarepo "github.com/smallbiznis/corebilling/internal/quota/repository/sqlc"
)

// SQLRepository implements Repository using sqlc-generated queries.
type SQLRepository struct {
	queries *quotarepo.Queries
}

// NewRepository constructs a quota repository backed by sqlc.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &SQLRepository{queries: quotarepo.New(pool)}
}

func (r *SQLRepository) GetQuotaLimit(ctx context.Context, tenantID string) (QuotaLimit, error) {
	rec, err := r.queries.GetQuotaLimit(ctx, tenantID)
	if err != nil {
		return QuotaLimit{}, err
	}
	threshold := rec.SoftWarningThreshold.Float64
	if !rec.SoftWarningThreshold.Valid {
		threshold = 0
	}
	return QuotaLimit{
		TenantID:             rec.TenantID,
		MaxEventsPerDay:      rec.MaxEventsPerDay,
		MaxUsageUnits:        rec.MaxUsageUnits,
		SoftWarningThreshold: threshold,
	}, nil
}

func (r *SQLRepository) GetQuotaUsage(ctx context.Context, tenantID string) (QuotaUsage, error) {
	rec, err := r.queries.GetQuotaUsage(ctx, tenantID)
	if err != nil {
		return QuotaUsage{}, err
	}
	return QuotaUsage{
		TenantID:    rec.TenantID,
		EventsToday: rec.EventsToday,
		UsageUnits:  rec.UsageUnits,
	}, nil
}

func (r *SQLRepository) UpsertQuotaUsage(ctx context.Context, params UpsertQuotaUsageParams) error {
	return r.queries.UpsertQuotaUsage(ctx, quotarepo.UpsertQuotaUsageParams{
		TenantID:    params.TenantID,
		EventsToday: params.EventsToday,
		UsageUnits:  params.UsageUnits,
	})
}

func (r *SQLRepository) ListTenantsOverLimit(ctx context.Context) ([]string, error) {
	rows, err := r.queries.ListTenantsOverLimit(ctx)
	if err != nil {
		return nil, err
	}
	tenants := make([]string, 0, len(rows))
	tenants = append(tenants, rows...)
	return tenants, nil
}

var _ Repository = (*SQLRepository)(nil)
