package domain

import "context"

// Repository provides pricing plan persistence.
type Repository interface {
	Create(ctx context.Context, plan PricingPlan) error
	GetByID(ctx context.Context, id string) (PricingPlan, error)
	List(ctx context.Context) ([]PricingPlan, error)
}
