package domain

import "context"

// Repository provides invoice persistence.
type Repository interface {
	Create(ctx context.Context, invoice Invoice) error
	GetByID(ctx context.Context, id string) (Invoice, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Invoice, error)
}
