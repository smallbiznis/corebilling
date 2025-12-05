package router

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
)

type noopBus struct{}

func (n *noopBus) Publish(ctx context.Context, envelopes ...events.EventEnvelope) error { return nil }
func (n *noopBus) Subscribe(ctx context.Context, subject string, group string, handler events.Handler) error {
	return nil
}
func (n *noopBus) Close() error { return nil }

func TestRouterBackpressure(t *testing.T) {
	t.Setenv("HANDLER_CONCURRENCY", "1")
	r := NewRouter(&noopBus{}, zap.NewNop(), "group", nil)

        block := make(chan struct{})
        started := make(chan struct{}, 2)
	handler := r.wrapHandler("subject", func(ctx context.Context, evt *events.Event) error {
		started <- struct{}{}
		<-block
		return nil
	})

	go func() {
		_ = handler(context.Background(), &events.Event{})
	}()
	<-started

	finished := make(chan struct{})
	go func() {
		_ = handler(context.Background(), &events.Event{})
		close(finished)
	}()

	select {
	case <-finished:
		t.Fatalf("second handler should have been blocked until release")
	case <-time.After(50 * time.Millisecond):
	}

	close(block)
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatalf("second handler did not complete after releasing block")
	}
}
