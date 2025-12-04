package domain

import "context"

// ListInvoicesFilter configures pagination and filters for invoice queries.
type ListInvoicesFilter struct {
	TenantID       string
	CustomerID     string
	SubscriptionID string
	Status         int32
	Limit          int
	Offset         int
}

// Repository provides invoice persistence.
type Repository interface {
	Create(ctx context.Context, invoice Invoice) error
	GetByID(ctx context.Context, id string) (Invoice, error)
	List(ctx context.Context, filter ListInvoicesFilter) ([]Invoice, bool, error)
}
