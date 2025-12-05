package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	billingcyclerepo "github.com/smallbiznis/corebilling/internal/billingcycle/repository/sqlc"
)

// SQLRepository implements Repository using sqlc-generated queries.
type SQLRepository struct {
	queries *billingcyclerepo.Queries
}

// NewRepository constructs a repository backed by sqlc queries.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &SQLRepository{queries: billingcyclerepo.New(pool)}
}

func (r *SQLRepository) GetCycleForTenant(ctx context.Context, tenantID string) (BillingCycle, error) {
	parsedTenantID, err := parseSnowflake(tenantID)
	if err != nil {
		return BillingCycle{}, err
	}
	rec, err := r.queries.GetCycleForTenant(ctx, parsedTenantID)
	if err != nil {
		return BillingCycle{}, err
	}
	var lastClosed *time.Time
	if rec.LastClosedAt.Valid {
		t := rec.LastClosedAt.Time
		lastClosed = &t
	}
	return BillingCycle{
		TenantID:     formatSnowflake(rec.TenantID),
		PeriodStart:  toTime(rec.PeriodStart),
		PeriodEnd:    toTime(rec.PeriodEnd),
		LastClosedAt: lastClosed,
	}, nil
}

func (r *SQLRepository) UpdateBillingCycle(ctx context.Context, params UpdateBillingCycleParams) error {
	tenantID, err := parseSnowflake(params.TenantID)
	if err != nil {
		return err
	}
	return r.queries.UpdateBillingCycle(ctx, billingcyclerepo.UpdateBillingCycleParams{
		TenantID:    tenantID,
		PeriodStart: toTimestamptz(params.PeriodStart),
		PeriodEnd:   toTimestamptz(params.PeriodEnd),
	})
}

func (r *SQLRepository) ListTenantsDueForCycleClose(ctx context.Context) ([]string, error) {
	rows, err := r.queries.ListTenantsDueForCycleClose(ctx)
	if err != nil {
		return nil, err
	}
	tenants := make([]string, 0, len(rows))
	for _, id := range rows {
		tenants = append(tenants, formatSnowflake(id))
	}
	return tenants, nil
}

var _ Repository = (*SQLRepository)(nil)

func toTime(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func parseSnowflake(value string) (int64, error) {
	if value == "" {
		return 0, errors.New("snowflake id required")
	}
	return strconv.ParseInt(value, 10, 64)
}

func formatSnowflake(value int64) string {
	return strconv.FormatInt(value, 10)
}
