package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/replay/repository/sqlc"
)

// NewRepository constructs a SQL-backed replay repository.
func NewRepository(db *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: sqlc.New(db)}
}

type sqlcRepository struct {
	queries *sqlc.Queries
}

func (r *sqlcRepository) GetEventByID(ctx context.Context, id string) (EventEnvelope, error) {
	parsedID, err := parseInt64(id)
	if err != nil {
		return EventEnvelope{}, err
	}
	evt, err := r.queries.GetEventByID(ctx, parsedID)
	if err != nil {
		return EventEnvelope{}, err
	}
	return toEnvelope(evt)
}

func (r *sqlcRepository) ListEventsForTenant(ctx context.Context, tenantID string) ([]EventEnvelope, error) {
	parsedTenantID, err := parseInt64(tenantID)
	if err != nil {
		return nil, err
	}
	events, err := r.queries.ListEventsForTenant(ctx, parsedTenantID)
	if err != nil {
		return nil, err
	}
	return mapEnvelopes(events)
}

func (r *sqlcRepository) ListEventsByType(ctx context.Context, eventType string) ([]EventEnvelope, error) {
	events, err := r.queries.ListEventsByType(ctx, pgtype.Text{
		String: eventType,
		Valid:  true,
	})
	if err != nil {
		return nil, err
	}
	return mapEnvelopes(events)
}

func (r *sqlcRepository) ListEventsByFilters(ctx context.Context, tenantID, eventType *string, since, until *time.Time) ([]EventEnvelope, error) {
	parsedTenantID, err := parseOptionalInt64(tenantID)
	if err != nil {
		return nil, err
	}

	typeParam := pgtype.Text{}
	if eventType != nil && *eventType != "" {
		typeParam = pgtype.Text{
			String: *eventType,
			Valid:  true,
		}
	}

	sinceTime := time.Time{}
	if since != nil {
		sinceTime = *since
	}
	untilTime := time.Now().UTC()
	if until != nil {
		untilTime = *until
	}

	events, err := r.queries.ListEventsByFilters(ctx, sqlc.ListEventsByFiltersParams{
		TenantID:    parsedTenantID,
		EventType:   typeParam,
		CreatedAt:   pgtype.Timestamptz{Time: sinceTime, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: untilTime, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapEnvelopes(events)
}

func mapEnvelopes(events []sqlc.BillingEvents) ([]EventEnvelope, error) {
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

func toEnvelope(evt sqlc.BillingEvents) (EventEnvelope, error) {
	if len(evt.Payload) == 0 {
		return EventEnvelope{}, errors.New("event payload missing")
	}

	decoded, err := events.UnmarshalEvent(evt.Payload)
	if err != nil {
		return EventEnvelope{}, err
	}

	if decoded.Id == "" {
		decoded.Id = formatInt64(evt.ID)
	}
	if decoded.Subject == "" {
		decoded.Subject = evt.Subject
	}
	if decoded.TenantId == "" {
		decoded.TenantId = formatInt64(evt.TenantID)
	}

	env := EventEnvelope{
		ID:        decoded.Id,
		Subject:   evt.Subject,
		TenantID:  formatInt64(evt.TenantID),
		Payload:   evt.Payload,
		CreatedAt: pgTimestamptz(evt.CreatedAt),
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

func parseInt64(value string) (int64, error) {
	if value == "" {
		return 0, errors.New("id required")
	}
	return strconv.ParseInt(value, 10, 64)
}

func parseOptionalInt64(value *string) (int64, error) {
	if value == nil || *value == "" {
		return 0, nil
	}
	return strconv.ParseInt(*value, 10, 64)
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func pgTimestamptz(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}

var _ Repository = (*sqlcRepository)(nil)
