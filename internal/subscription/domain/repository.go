package domain

import "context"

// ListSubscriptionsFilter configures pagination for tenant subscriptions.
type ListSubscriptionsFilter struct {
	TenantID   string
	CustomerID string
	Limit      int
	Offset     int
}

// Repository provides subscription persistence.
type Repository interface {
	Create(ctx context.Context, sub Subscription) error
	GetByID(ctx context.Context, id string) (Subscription, error)
	List(ctx context.Context, filter ListSubscriptionsFilter) ([]Subscription, bool, error)
	Update(ctx context.Context, sub Subscription) error
}
