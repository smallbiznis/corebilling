package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/smallbiznis/corebilling/internal/events"
	replayrepo "github.com/smallbiznis/corebilling/internal/events/replay/repository/sqlc"
)

// NewRepository constructs a SQL-backed replay repository.
func NewRepository(db *sql.DB) Repository {
	return &sqlcRepository{queries: replayrepo.New(db)}
}

type sqlcRepository struct {
	queries *replayrepo.Queries
}

func (r *sqlcRepository) GetEventByID(ctx context.Context, id string) (EventEnvelope, error) {
	evt, err := r.queries.GetEventByID(ctx, id)
	if err != nil {
		return EventEnvelope{}, err
	}
	return toEnvelope(evt)
}

func (r *sqlcRepository) ListEventsForTenant(ctx context.Context, tenantID string) ([]EventEnvelope, error) {
	events, err := r.queries.ListEventsForTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return mapEnvelopes(events)
}

func (r *sqlcRepository) ListEventsByType(ctx context.Context, eventType string) ([]EventEnvelope, error) {
	events, err := r.queries.ListEventsByType(ctx, eventType)
	if err != nil {
		return nil, err
	}
	return mapEnvelopes(events)
}

func (r *sqlcRepository) ListEventsByFilters(ctx context.Context, tenantID, eventType *string, since, until *time.Time) ([]EventEnvelope, error) {
	events, err := r.queries.ListEventsByFilters(ctx, tenantID, eventType, since, until)
	if err != nil {
		return nil, err
	}
	return mapEnvelopes(events)
}

func mapEnvelopes(events []replayrepo.BillingEvent) ([]EventEnvelope, error) {
	envelopes := make([]EventEnvelope, 0, len(events))
	for _, evt := range events {
		env, err := toEnvelope(evt)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	return envelopes, nil
}

func toEnvelope(evt replayrepo.BillingEvent) (EventEnvelope, error) {
	if len(evt.Payload) == 0 {
		return EventEnvelope{}, errors.New("event payload missing")
	}

	decoded, err := events.UnmarshalEvent(evt.Payload)
	if err != nil {
		return EventEnvelope{}, err
	}

	if decoded.Id == "" {
		decoded.Id = evt.ID
	}
	if decoded.Subject == "" {
		decoded.Subject = evt.Subject
	}
	if decoded.TenantId == "" {
		decoded.TenantId = evt.TenantID
	}

	env := EventEnvelope{
		ID:        decoded.Id,
		Subject:   evt.Subject,
		TenantID:  evt.TenantID,
		Payload:   evt.Payload,
		CreatedAt: evt.CreatedAt,
		Event:     decoded,
	}

	if evt.ResourceID.Valid {
		env.ResourceID = evt.ResourceID.String
	}
	if evt.EventType.Valid {
		env.EventType = evt.EventType.String
	}

	return env, nil
}

var _ Repository = (*sqlcRepository)(nil)
