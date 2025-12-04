package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Repository implements OutboxRepository using pgx and JSONB metadata fields.
type Repository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
	logger *zap.Logger
}

// NewRepository constructs a new outbox repository.
func NewRepository(pool *pgxpool.Pool, logger *zap.Logger) OutboxRepository {
	return &Repository{pool: pool, tracer: otel.Tracer("events.outbox"), logger: logger}
}

// InsertOutboxEvent inserts a new event into billing_events with pending status.
func (r *Repository) InsertOutboxEvent(ctx context.Context, evt *OutboxEvent) error {
	if evt == nil || evt.Event == nil {
		return errors.New("event payload required")
	}

	ctx, cid := correlation.EnsureCorrelationID(ctx)
	ctx, span := r.tracer.Start(ctx, "outbox.write")
	defer span.End()

	log := ctxlogger.FromContext(ctx)
	log.Info("outbox.write", zap.String("subject", evt.Subject))

	if evt.Event.Id == "" {
		evt.Event.Id = uuid.NewString()
	}
	if evt.Event.Subject == "" {
		evt.Event.Subject = evt.Subject
	}
	if evt.Event.TenantId == "" {
		evt.Event.TenantId = evt.TenantID
	}

	if evt.Event.Metadata == nil {
		evt.Event.Metadata = &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	if evt.Event.Metadata.Fields == nil {
		evt.Event.Metadata.Fields = map[string]*structpb.Value{}
	}
	evt.Event.Metadata.Fields["correlation_id"] = structpb.NewStringValue(cid)

	correlation.InjectTraceIntoEvent(evt.Event, span)
	span.SetAttributes(
		attribute.String("event.subject", evt.Event.Subject),
		attribute.String("event.correlation_id", cid),
		attribute.String("event.trace_id", span.SpanContext().TraceID().String()),
		attribute.String("outbox.table", "billing_events"),
	)

	now := time.Now().UTC()
	if evt.Event.CreatedAt == nil {
		evt.Event.CreatedAt = timestamppb.New(now)
	}

	if _, err := ApplyMetadata(evt.Event, OutboxStatusPending, 0, now, ""); err != nil {
		return err
	}

	payload, err := marshalEvent(evt.Event)
	if err != nil {
		return err
	}

	_, err = r.exec(ctx,
		`INSERT INTO billing_events (id, subject, tenant_id, resource_id, payload, created_at) VALUES ($1,$2,$3,$4,$5,now())`,
		evt.Event.Id, evt.Event.Subject, evt.Event.TenantId, nullIfEmpty(evt.ResourceID), payload,
	)
	return err
}

// FetchPendingEvents retrieves events eligible for dispatch.
func (r *Repository) FetchPendingEvents(ctx context.Context, limit int32, now time.Time) ([]OutboxEvent, error) {
	rows, err := r.query(ctx, `
        SELECT 
					id,
					subject,
					tenant_id,
					resource_id,
					payload,
					status,
					retry_count,
					next_attempt_at,
					last_error,
					created_at
			FROM billing_events
			WHERE status = 'pending'
				AND (next_attempt_at IS NULL OR next_attempt_at <= $1)
			ORDER BY id
			LIMIT $2;`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventsOut []OutboxEvent
	for rows.Next() {
		var evt OutboxEvent
		if err := rows.Scan(&evt.ID, &evt.Subject, &evt.TenantID, &evt.ResourceID, &evt.Payload, &evt.CreatedAt); err != nil {
			return nil, err
		}
		parsed, err := unmarshalEvent(evt.Payload)
		if err != nil {
			return nil, err
		}
		evt.Event = parsed
		status, retry, nextAttempt, lastErr := ExtractMetadata(parsed)
		evt.Status = status
		evt.RetryCount = retry
		evt.NextAttemptAt = nextAttempt
		evt.LastError = lastErr
		eventsOut = append(eventsOut, evt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return eventsOut, nil
}

// MarkDispatched marks an event as dispatched within metadata.
func (r *Repository) MarkDispatched(ctx context.Context, id string) error {
	_, err := r.exec(ctx, `
        UPDATE billing_events
        SET payload = jsonb_set(payload, '{metadata,outbox_status}', to_jsonb($2::text), true)
        WHERE id = $1`, id, string(OutboxStatusDispatched))
	return err
}

// MarkFailed updates retry metadata and schedules the next attempt.
func (r *Repository) MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, lastError string, retryCount int32) error {
	_, err := r.exec(ctx, `
        UPDATE billing_events
        SET payload = jsonb_set(
            jsonb_set(
                jsonb_set(
                    jsonb_set(payload, '{metadata,outbox_status}', to_jsonb($2::text), true),
                    '{metadata,outbox_retry_count}', to_jsonb($3::int), true),
                '{metadata,outbox_next_attempt_at}', to_jsonb($4::text), true),
            '{metadata,outbox_last_error}', to_jsonb($5::text), true)
        WHERE id = $1`, id, string(OutboxStatusFailed), retryCount, nextAttemptAt.UTC().Format(time.RFC3339), lastError)
	return err
}

// MoveToDeadLetter moves an event to DLQ status.
func (r *Repository) MoveToDeadLetter(ctx context.Context, id string, lastError string) error {
	_, err := r.exec(ctx, `
        UPDATE billing_events
        SET payload = jsonb_set(
            jsonb_set(payload, '{metadata,outbox_status}', to_jsonb($2::text), true),
            '{metadata,outbox_last_error}', to_jsonb($3::text), true)
        WHERE id = $1`, id, string(OutboxStatusDeadLetter), lastError)
	return err
}

func (r *Repository) exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return r.pool.Exec(ctx, sql, args...)
}

func (r *Repository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return r.pool.Query(ctx, sql, args...)
}

func marshalEvent(evt *eventv1.Event) ([]byte, error) {
	marshaler := protojson.MarshalOptions{UseEnumNumbers: false, EmitUnpopulated: false}
	return marshaler.Marshal(evt)
}

func unmarshalEvent(data []byte) (*eventv1.Event, error) {
	var evt eventv1.Event
	if len(data) == 0 {
		return &evt, nil
	}
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(data, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}

func nullIfEmpty(val string) any {
	if val == "" {
		return nil
	}
	return val
}

var _ OutboxRepository = (*Repository)(nil)
