package outbox

import (
	"context"
	"time"
)

// OutboxRepository abstracts persistence for billing_events as an outbox.
type OutboxRepository interface {
	InsertOutboxEvent(ctx context.Context, evt *OutboxEvent) error
	FetchPendingEvents(ctx context.Context, limit int32, now time.Time) ([]OutboxEvent, error)
	MarkDispatched(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, lastError string, retryCount int32) error
	MoveToDeadLetter(ctx context.Context, id string, lastError string) error
}
