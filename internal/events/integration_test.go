package events_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	"github.com/smallbiznis/corebilling/internal/events/router"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventLifecyclePublishesFollowUps(t *testing.T) {
	ctx := context.Background()
	bus := newTestBus()
	router := router.NewRouter(bus, zap.NewNop(), "integration-test", nil)

	fooCh := make(chan string, 1)
	barCh := make(chan string, 1)

	router.Register("subject.foo", func(ctx context.Context, evt *events.Event) error {
		fooCh <- evt.GetId()
		followUp := &eventv1.Event{
			Id:        "followup-1",
			Subject:   "subject.bar",
			TenantId:  evt.GetTenantId(),
			CreatedAt: timestamppb.Now(),
		}
		env := events.NewEnvelope(followUp)
		if _, err := env.Prepare(ctx); err != nil {
			return err
		}
		return bus.Publish(ctx, env)
	})
	router.Register("subject.bar", func(ctx context.Context, evt *events.Event) error {
		barCh <- evt.GetSubject()
		return nil
	})
	if err := router.Start(ctx); err != nil {
		t.Fatalf("failed to start router: %v", err)
	}

	repo := newInMemoryOutbox()
	dispatcher := outbox.NewDispatcher(repo, bus, zap.NewNop(), nil)

	event := &eventv1.Event{
		Id:        "evt-1",
		Subject:   "subject.foo",
		TenantId:  "tenant-123",
		CreatedAt: timestamppb.Now(),
	}
	if err := repo.InsertOutboxEvent(ctx, &outbox.OutboxEvent{
		ID:       "evt-1",
		Subject:  "subject.foo",
		TenantID: "tenant-123",
		Event:    event,
		Status:   outbox.OutboxStatusPending,
	}); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	dispatcher.Dispatch(ctx)

	select {
	case id := <-fooCh:
		if id != "evt-1" {
			t.Fatalf("unexpected subject.foo event id: %s", id)
		}
	case <-time.After(time.Second):
		t.Fatalf("subject.foo handler not invoked")
	}
	select {
	case subj := <-barCh:
		if subj != "subject.bar" {
			t.Fatalf("unexpected follow-up subject: %s", subj)
		}
	case <-time.After(time.Second):
		t.Fatalf("subject.bar handler not invoked")
	}

	if pending, err := repo.FetchPendingEvents(ctx, 10, time.Now()); err != nil {
		t.Fatalf("checking pending events: %v", err)
	} else if len(pending) != 0 {
		t.Fatalf("expected no pending events, got %d", len(pending))
	}
}

type testBus struct {
	mu        sync.Mutex
	handlers  map[string][]events.Handler
	published []events.EventEnvelope
}

func newTestBus() *testBus {
	return &testBus{
		handlers: make(map[string][]events.Handler),
	}
}

func (b *testBus) Publish(ctx context.Context, envelopes ...events.EventEnvelope) error {
	b.mu.Lock()
	b.published = append(b.published, envelopes...)
	b.mu.Unlock()

	for _, env := range envelopes {
		b.mu.Lock()
		handlerList := append([]events.Handler{}, b.handlers[env.Subject]...)
		b.mu.Unlock()
		for _, handler := range handlerList {
			if err := handler(ctx, env.Event); err != nil {
				return fmt.Errorf("handler error: %w", err)
			}
		}
	}
	return nil
}

func (b *testBus) Subscribe(ctx context.Context, subject, group string, handler events.Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[subject] = append(b.handlers[subject], handler)
	return nil
}

func (b *testBus) Close() error {
	return nil
}

type inMemoryOutbox struct {
	mu     sync.Mutex
	events map[string]outbox.OutboxEvent
	order  []string
}

func newInMemoryOutbox() *inMemoryOutbox {
	return &inMemoryOutbox{
		events: make(map[string]outbox.OutboxEvent),
	}
}

func (o *inMemoryOutbox) InsertOutboxEvent(ctx context.Context, evt *outbox.OutboxEvent) error {
	if evt == nil || evt.ID == "" {
		return fmt.Errorf("event id required")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, exists := o.events[evt.ID]; exists {
		return fmt.Errorf("event %s already exists", evt.ID)
	}
	o.order = append(o.order, evt.ID)
	o.events[evt.ID] = *evt
	return nil
}

func (o *inMemoryOutbox) FetchPendingEvents(ctx context.Context, limit int32, now time.Time) ([]outbox.OutboxEvent, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	var list []outbox.OutboxEvent
	for _, id := range o.order {
		if len(list) >= int(limit) {
			break
		}
		list = append(list, o.events[id])
	}
	return list, nil
}

func (o *inMemoryOutbox) MarkDispatched(ctx context.Context, id string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, ok := o.events[id]; !ok {
		return fmt.Errorf("event %s not found", id)
	}
	delete(o.events, id)
	for i, stored := range o.order {
		if stored == id {
			o.order = append(o.order[:i], o.order[i+1:]...)
			break
		}
	}
	return nil
}

func (o *inMemoryOutbox) MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, lastError string, retryCount int32) error {
	return nil
}

func (o *inMemoryOutbox) MoveToDeadLetter(ctx context.Context, id, lastError string) error {
	return nil
}
