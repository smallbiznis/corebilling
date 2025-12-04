package domain

import (
	"context"
)

// Repository defines persistence for invoice engine runs.
type Repository interface {
	Create(ctx context.Context, run Run) error
}
