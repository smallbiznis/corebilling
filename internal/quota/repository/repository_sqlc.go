package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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
	parsedTenantID, err := parseSnowflake(tenantID)
	if err != nil {
		return QuotaLimit{}, err
	}
	rec, err := r.queries.GetQuotaLimit(ctx, parsedTenantID)
	if err != nil {
		return QuotaLimit{}, err
	}
	threshold := rec.SoftWarningThreshold.Float64
	if !rec.SoftWarningThreshold.Valid {
		threshold = 0
	}
	return QuotaLimit{
		TenantID:             formatSnowflake(rec.TenantID),
		MaxEventsPerDay:      rec.MaxEventsPerDay,
		MaxUsageUnits:        rec.MaxUsageUnits,
		SoftWarningThreshold: threshold,
	}, nil
}

func (r *SQLRepository) GetQuotaUsage(ctx context.Context, tenantID string) (QuotaUsage, error) {
	parsedTenantID, err := parseSnowflake(tenantID)
	if err != nil {
		return QuotaUsage{}, err
	}
	rec, err := r.queries.GetQuotaUsage(ctx, parsedTenantID)
	if err != nil {
		return QuotaUsage{}, err
	}
	tenantIDStr, err := formatSnowflakeFromInterface(rec.TenantID)
	if err != nil {
		return QuotaUsage{}, err
	}
	return QuotaUsage{
		TenantID:    tenantIDStr,
		EventsToday: rec.EventsToday,
		UsageUnits:  rec.UsageUnits,
	}, nil
}

func (r *SQLRepository) UpsertQuotaUsage(ctx context.Context, params UpsertQuotaUsageParams) error {
	tenantID, err := parseSnowflake(params.TenantID)
	if err != nil {
		return err
	}
	return r.queries.UpsertQuotaUsage(ctx, quotarepo.UpsertQuotaUsageParams{
		TenantID:    tenantID,
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
	for _, row := range rows {
		id, err := formatSnowflakeFromInterface(row)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, id)
	}
	return tenants, nil
}

var _ Repository = (*SQLRepository)(nil)

func parseSnowflake(value string) (int64, error) {
	if value == "" {
		return 0, errors.New("snowflake id required")
	}
	return strconv.ParseInt(value, 10, 64)
}

func formatSnowflake(value int64) string {
	return strconv.FormatInt(value, 10)
}

func formatSnowflakeFromInterface(value interface{}) (string, error) {
	switch v := value.(type) {
	case int64:
		return formatSnowflake(v), nil
	case int32:
		return formatSnowflake(int64(v)), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported tenant id type %T", v)
	}
}
