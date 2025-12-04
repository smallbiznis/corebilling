package domain

import "context"

// Repository defines persistence methods for billing records.
type Repository interface {
	Create(ctx context.Context, record BillingRecord) error
	GetByID(ctx context.Context, id string) (BillingRecord, error)
	ListByTenant(ctx context.Context, tenantID string) ([]BillingRecord, error)
}
