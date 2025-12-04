package domain

import "context"

// Repository provides subscription persistence.
type Repository interface {
	Create(ctx context.Context, sub Subscription) error
	GetByID(ctx context.Context, id string) (Subscription, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Subscription, error)
}
