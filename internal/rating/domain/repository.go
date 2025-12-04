package domain

import "context"

// Repository defines rating persistence.
type Repository interface {
	Create(ctx context.Context, rating RatingResult) error
	GetByUsage(ctx context.Context, usageID string) ([]RatingResult, error)
}
