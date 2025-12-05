package outbox

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestTenantFailureTrackerThrottles(t *testing.T) {
	tracker := newTenantFailureTracker(10*time.Second, 45*time.Second, 3)
	tenant := "tenant-1"
	base := time.Now().Add(5 * time.Second)

	next := base
	for i := 0; i < 3; i++ {
		next = tracker.nextAttempt(tenant, base)
	}

	if !next.After(base) {
		t.Fatalf("expected throttled time after base; got %v <= %v", next, base)
	}
}

func TestDispatcherFailureTracking(t *testing.T) {
	repo := &stubOutboxRepo{events: map[string]OutboxEvent{}}
	dispatcher := NewDispatcher(repo, nil, zap.NewNop(), nil)
	dispatcher.failures = newTenantFailureTracker(time.Minute, time.Hour, 1)

	evt := OutboxEvent{ID: "1", TenantID: "tenant-a"}
	repo.events[evt.ID] = evt
	ctx := context.Background()

	before := ComputeNextAttempt(1, dispatcher.baseBackoff, dispatcher.maxBackoff)
	dispatcher.handleFailure(ctx, evt, assertError("boom"))
	after := repo.events[evt.ID]
	if after.NextAttemptAt.Before(before) {
		t.Fatalf("expected throttled next attempt >= %v, got %v", before, after.NextAttemptAt)
	}
}

type testError string

func (e testError) Error() string { return string(e) }

func assertError(msg string) error { return testError(msg) }

type stubOutboxRepo struct {
	events map[string]OutboxEvent
}

func (s *stubOutboxRepo) InsertOutboxEvent(ctx context.Context, evt *OutboxEvent) error {
	s.events[evt.ID] = *evt
	return nil
}

func (s *stubOutboxRepo) FetchPendingEvents(ctx context.Context, limit int32, now time.Time) ([]OutboxEvent, error) {
	return nil, nil
}

func (s *stubOutboxRepo) MarkDispatched(ctx context.Context, id string) error { return nil }

func (s *stubOutboxRepo) MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, lastError string, retryCount int32) error {
	evt := s.events[id]
	evt.NextAttemptAt = nextAttemptAt
	evt.RetryCount = retryCount
	evt.LastError = lastError
	s.events[id] = evt
	return nil
}

func (s *stubOutboxRepo) MoveToDeadLetter(ctx context.Context, id string, lastError string) error {
	evt := s.events[id]
	evt.Status = OutboxStatusDeadLetter
	evt.LastError = lastError
	s.events[id] = evt
	return nil
}
