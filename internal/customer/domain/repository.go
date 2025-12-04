package domain

import "context"

// Repository defines persistence for customers.
type Repository interface {
	Create(ctx context.Context, customer Customer) error
	GetByID(ctx context.Context, id string) (Customer, error)
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]Customer, error)
	Update(ctx context.Context, customer Customer) error
}
