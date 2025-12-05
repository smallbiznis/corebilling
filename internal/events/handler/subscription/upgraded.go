package subscription

import (
	"context"
	"errors"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/handler"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	subdomain "github.com/smallbiznis/corebilling/internal/subscription/domain"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// SubscriptionUpgradedHandler handles subscription.upgraded events.
type SubscriptionUpgradedHandler struct {
	svc       *subdomain.Service
	tracker   *outbox.IdempotencyTracker
	publisher events.Publisher
	logger    *zap.Logger
}

// NewSubscriptionUpgradedHandler constructs the handler.
func NewSubscriptionUpgradedHandler(
	svc *subdomain.Service,
	publisher events.Publisher,
	tracker *outbox.IdempotencyTracker,
	logger *zap.Logger,
) handler.HandlerOut {
	return handler.HandlerOut{
		Handler: &SubscriptionUpgradedHandler{
			svc:       svc,
			publisher: publisher,
			tracker:   tracker,
			logger:    logger.Named("subscription.upgraded"),
		},
	}
}

func (h *SubscriptionUpgradedHandler) Subject() string {
	return "subscription.upgraded"
}

func (h *SubscriptionUpgradedHandler) Handle(ctx context.Context, evt *events.Event) error {
	if evt == nil {
		return errors.New("event required")
	}
	if h.tracker != nil && h.tracker.SeenBefore(evt.GetId()) {
		h.logger.Debug("event already processed", zap.String("event_id", evt.GetId()))
		return nil
	}

	subID := handler.ParseString(evt.GetData(), "subscription_id")
	if subID == "" {
		return errors.New("subscription_id required")
	}
	sub, err := h.svc.Get(ctx, subID)
	if err != nil {
		return err
	}
	newPrice := handler.ParseString(evt.GetData(), "price_id")
	if newPrice != "" {
		sub.PriceID = newPrice
	}
	if err := h.svc.Update(ctx, sub); err != nil {
		return err
	}

	if h.publisher != nil {
		payload := map[string]*structpb.Value{
			"subscription_id": structpb.NewStringValue(sub.ID),
			"price_id":        structpb.NewStringValue(sub.PriceID),
		}
		if child, childErr := handler.NewFollowUpEvent(evt, "subscription.price.updated", sub.TenantID, payload); childErr == nil {
			_ = h.publisher.Publish(ctx, events.EventEnvelope{Event: child})
		}
	}
	return nil
}
