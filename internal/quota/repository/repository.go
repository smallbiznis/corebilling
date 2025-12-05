package repository

import "context"

type QuotaLimit struct {
	TenantID             string
	MaxEventsPerDay      int64
	MaxUsageUnits        int64
	SoftWarningThreshold float64
}

type QuotaUsage struct {
	TenantID    string
	EventsToday int64
	UsageUnits  int64
}

type UpsertQuotaUsageParams struct {
	TenantID    string
	EventsToday int64
	UsageUnits  int64
}

type Repository interface {
	GetQuotaLimit(ctx context.Context, tenantID string) (QuotaLimit, error)
	GetQuotaUsage(ctx context.Context, tenantID string) (QuotaUsage, error)
	UpsertQuotaUsage(ctx context.Context, params UpsertQuotaUsageParams) error
	ListTenantsOverLimit(ctx context.Context) ([]string, error)
}
