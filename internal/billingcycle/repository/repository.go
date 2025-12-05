package repository

import (
	"context"
	"time"
)

type BillingCycle struct {
	TenantID     string
	PeriodStart  time.Time
	PeriodEnd    time.Time
	LastClosedAt *time.Time
}

type UpdateBillingCycleParams struct {
	TenantID    string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type Repository interface {
	GetCycleForTenant(ctx context.Context, tenantID string) (BillingCycle, error)
	UpdateBillingCycle(ctx context.Context, params UpdateBillingCycleParams) error
	ListTenantsDueForCycleClose(ctx context.Context) ([]string, error)
}
