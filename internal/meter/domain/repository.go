package domain

import "context"

// Repository defines persistence for meters.
type Repository interface {
	Create(ctx context.Context, meter Meter) error
	GetByID(ctx context.Context, id string) (Meter, error)
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]Meter, error)
	Update(ctx context.Context, meter Meter) error
}
