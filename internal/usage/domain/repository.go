package domain

import (
	"context"
	"time"
)

// ListUsageFilter configures filters for listing usage records.
type ListUsageFilter struct {
	TenantID       string
	SubscriptionID string
	CustomerID     string
	MeterCode      string
	From           time.Time
	To             time.Time
	Limit          int
	Offset         int
}

// Repository for usage records.
type Repository interface {
	Create(ctx context.Context, usage UsageRecord) error
	List(ctx context.Context, filter ListUsageFilter) ([]UsageRecord, bool, error)
}
