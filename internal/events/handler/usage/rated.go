package usage

import (
	"context"
	"errors"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/handler"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// UsageRatedHandler handles usage.rated events and forwards billing.
type UsageRatedHandler struct {
	tracker   *outbox.IdempotencyTracker
	publisher events.Publisher
	logger    *zap.Logger
}

// NewUsageRatedHandler constructs the handler.
func NewUsageRatedHandler(
	publisher events.Publisher,
	tracker *outbox.IdempotencyTracker,
	logger *zap.Logger,
) handler.HandlerOut {
	return handler.HandlerOut{
		Handler: &UsageRatedHandler{
			tracker:   tracker,
			publisher: publisher,
			logger:    logger.Named("usage.rated"),
		},
	}
}

func (h *UsageRatedHandler) Subject() string {
	return "usage.rated"
}

func (h *UsageRatedHandler) Handle(ctx context.Context, evt *events.Event) error {
	if evt == nil {
		return errors.New("event required")
	}
	if h.tracker != nil && h.tracker.SeenBefore(evt.GetId()) {
		h.logger.Debug("event already processed", zap.String("event_id", evt.GetId()))
		return nil
	}
	if h.publisher == nil {
		return nil
	}

	data := evt.GetData()
	subscriptionID := handler.ParseString(data, "subscription_id")
	if subscriptionID == "" {
		return errors.New("subscription_id required")
	}
	amount := handler.ParseFloat(data, "amount_cents")
	payload := map[string]*structpb.Value{
		"subscription_id": structpb.NewStringValue(subscriptionID),
		"amount_cents":    structpb.NewNumberValue(amount),
	}
	if child, childErr := handler.NewFollowUpEvent(evt, "invoice.generated", evt.GetTenantId(), payload); childErr == nil {
		return h.publisher.Publish(ctx, events.EventEnvelope{Event: child})
	}
	return nil
}
