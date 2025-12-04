package domain

import "context"

// Repository for usage records.
type Repository interface {
	Create(ctx context.Context, usage UsageRecord) error
	ListBySubscription(ctx context.Context, subscriptionID string) ([]UsageRecord, error)
}
