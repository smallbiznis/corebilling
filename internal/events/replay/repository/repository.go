package repository

import (
	"context"
	"time"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
)

// EventEnvelope represents a stored billing event with decoded payload.
type EventEnvelope struct {
	ID         string
	Subject    string
	TenantID   string
	ResourceID string
	EventType  string
	Payload    []byte
	CreatedAt  time.Time
	Event      *eventv1.Event
}

// Repository exposes read operations for stored events.
type Repository interface {
	GetEventByID(ctx context.Context, id string) (EventEnvelope, error)
	ListEventsForTenant(ctx context.Context, tenantID string) ([]EventEnvelope, error)
	ListEventsByType(ctx context.Context, eventType string) ([]EventEnvelope, error)
	ListEventsByFilters(ctx context.Context, tenantID, eventType *string, since, until *time.Time) ([]EventEnvelope, error)
}
