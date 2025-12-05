package replayrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// BillingEvent mirrors the billing_events table.
type BillingEvent struct {
	ID         string
	Subject    string
	TenantID   string
	ResourceID sql.NullString
	EventType  sql.NullString
	Payload    []byte
	CreatedAt  time.Time
}

// Queries provides accessors for replay queries.
type Queries struct {
	db *sql.DB
}

// New constructs a Queries instance.
func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

// GetEventByID returns an event by identifier.
func (q *Queries) GetEventByID(ctx context.Context, id string) (BillingEvent, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT * FROM billing_events WHERE id = $1 LIMIT 1`, id)
	if err != nil {
		return BillingEvent{}, err
	}
	events, err := scanBillingEvents(rows)
	if err != nil {
		return BillingEvent{}, err
	}
	if len(events) == 0 {
		return BillingEvent{}, sql.ErrNoRows
	}
	return events[0], nil
}

// ListEventsForTenant returns events for a tenant ordered by creation time.
func (q *Queries) ListEventsForTenant(ctx context.Context, tenantID string) ([]BillingEvent, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT * FROM billing_events WHERE tenant_id = $1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	return scanBillingEvents(rows)
}

// ListEventsByType returns events matching an event type ordered by creation time.
func (q *Queries) ListEventsByType(ctx context.Context, eventType string) ([]BillingEvent, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT * FROM billing_events WHERE event_type = $1 ORDER BY created_at ASC`, eventType)
	if err != nil {
		return nil, err
	}
	return scanBillingEvents(rows)
}

// ListEventsByFilters applies optional filters for tenant, type, and time range.
func (q *Queries) ListEventsByFilters(ctx context.Context, tenantID, eventType *string, since, until *time.Time) ([]BillingEvent, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT *
FROM billing_events
WHERE ($1::TEXT IS NULL OR tenant_id = $1)
  AND ($2::TEXT IS NULL OR event_type = $2)
  AND ($3::TIMESTAMPTZ IS NULL OR created_at >= $3)
  AND ($4::TIMESTAMPTZ IS NULL OR created_at <= $4)
ORDER BY created_at ASC`, tenantID, eventType, since, until)
	if err != nil {
		return nil, err
	}
	return scanBillingEvents(rows)
}

func scanBillingEvents(rows *sql.Rows) ([]BillingEvent, error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var events []BillingEvent
	for rows.Next() {
		values := make([]any, len(cols))
		for i := range values {
			values[i] = new(any)
		}
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		evt := BillingEvent{}
		for idx, col := range cols {
			val := *(values[idx].(*any))
			switch col {
			case "id":
				evt.ID = toString(val)
			case "subject":
				evt.Subject = toString(val)
			case "tenant_id":
				evt.TenantID = toString(val)
			case "resource_id":
				evt.ResourceID = toNullString(val)
			case "event_type":
				evt.EventType = toNullString(val)
			case "payload":
				switch v := val.(type) {
				case []byte:
					evt.Payload = append([]byte(nil), v...)
				case string:
					evt.Payload = []byte(v)
				}
			case "created_at":
				evt.CreatedAt = toTime(val)
			}
		}
		events = append(events, evt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func toString(val any) string {
	switch v := val.(type) {
	case nil:
		return ""
	case string:
		return v
	case []byte:
		return string(v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	}
	return fmt.Sprint(val)
}

func toNullString(val any) sql.NullString {
	s := toString(val)
	return sql.NullString{String: s, Valid: s != ""}
}

func toTime(val any) time.Time {
	switch v := val.(type) {
	case time.Time:
		return v
	case []byte:
		if parsed, err := time.Parse(time.RFC3339Nano, string(v)); err == nil {
			return parsed
		}
		if parsed, err := time.Parse("2006-01-02 15:04:05Z07:00", string(v)); err == nil {
			return parsed
		}
	case string:
		if parsed, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return parsed
		}
		if parsed, err := time.Parse("2006-01-02 15:04:05Z07:00", v); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
