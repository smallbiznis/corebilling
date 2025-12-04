package domain

import "context"

// Repository defines persistence methods for billing runs.
type Repository interface {
    Create(ctx context.Context, run BillingRun) error
    GetByID(ctx context.Context, id int64) (BillingRun, error)
    ListBySubscription(ctx context.Context, subscriptionID int64) ([]BillingRun, error)
}
