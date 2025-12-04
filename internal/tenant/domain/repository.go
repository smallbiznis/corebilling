package domain

import "context"

// Repository defines persistence for tenants.
type Repository interface {
	Create(ctx context.Context, tenant Tenant) error
	GetByID(ctx context.Context, id string) (Tenant, error)
	List(ctx context.Context, filter ListFilter) ([]Tenant, error)
	Update(ctx context.Context, tenant Tenant) error
}
