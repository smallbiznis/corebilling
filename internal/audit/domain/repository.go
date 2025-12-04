package domain

import "context"

// Repository defines persistence for audit events.
type Repository interface {
	Create(ctx context.Context, event AuditEvent) error
	GetByID(ctx context.Context, id string) (AuditEvent, error)
	List(ctx context.Context, filter ListFilter) ([]AuditEvent, error)
}
