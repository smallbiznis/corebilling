package subscription

import (
	"context"
	"errors"
	"time"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/handler"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	subdomain "github.com/smallbiznis/corebilling/internal/subscription/domain"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// SubscriptionCreatedHandler handles subscription.created events.
type SubscriptionCreatedHandler struct {
	svc       *subdomain.Service
	tracker   *outbox.IdempotencyTracker
	publisher events.Publisher
	logger    *zap.Logger
}

// NewSubscriptionCreatedHandler constructs the handler.
func NewSubscriptionCreatedHandler(
	svc *subdomain.Service,
	publisher events.Publisher,
	tracker *outbox.IdempotencyTracker,
	logger *zap.Logger,
) handler.HandlerOut {
	return handler.HandlerOut{
		Handler: &SubscriptionCreatedHandler{
			svc:       svc,
			publisher: publisher,
			tracker:   tracker,
			logger:    logger.Named("subscription.created"),
		},
	}
}

func (h *SubscriptionCreatedHandler) Subject() string {
	return "subscription.created"
}

func (h *SubscriptionCreatedHandler) Handle(ctx context.Context, evt *events.Event) error {
	if evt == nil {
		return errors.New("event required")
	}
	if h.tracker != nil && h.tracker.SeenBefore(evt.GetId()) {
		h.logger.Debug("event already processed", zap.String("event_id", evt.GetId()))
		return nil
	}

	sub, err := h.buildSubscription(evt)
	if err != nil {
		return err
	}
	if err := h.svc.Create(ctx, sub); err != nil {
		return err
	}
	if h.publisher != nil {
		payload := map[string]*structpb.Value{
			"subscription_id": structpb.NewStringValue(sub.ID),
			"customer_id":     structpb.NewStringValue(sub.CustomerID),
		}
		if child, childErr := handler.NewFollowUpEvent(evt, "subscription.provisioned", sub.TenantID, payload); childErr == nil {
			_ = h.publisher.Publish(ctx, events.EventEnvelope{Event: child})
		}
	}
	return nil
}

func (h *SubscriptionCreatedHandler) buildSubscription(evt *events.Event) (subdomain.Subscription, error) {
	data := evt.GetData()
	trialStart, err := handler.ParseTime(data, "trial_start_at")
	if err != nil {
		return subdomain.Subscription{}, err
	}
	trialEnd, err := handler.ParseTime(data, "trial_end_at")
	if err != nil {
		return subdomain.Subscription{}, err
	}
	cancelAt, err := handler.ParseTime(data, "cancel_at")
	if err != nil {
		return subdomain.Subscription{}, err
	}
	status := int32(handler.ParseFloat(data, "status"))
	autoRenew := handler.ParseBool(data, "auto_renew")
	startAtVal, err := handler.ParseTime(data, "start_at")
	if err != nil {
		return subdomain.Subscription{}, err
	}
	currentStartVal, err := handler.ParseTime(data, "current_period_start")
	if err != nil {
		return subdomain.Subscription{}, err
	}
	currentEndVal, err := handler.ParseTime(data, "current_period_end")
	if err != nil {
		return subdomain.Subscription{}, err
	}
	startAt := timeOrZero(startAtVal)
	currentStart := timeOrZero(currentStartVal)
	currentEnd := timeOrZero(currentEndVal)

	return subdomain.Subscription{
		ID:                 handler.ParseString(data, "id"),
		TenantID:           evt.GetTenantId(),
		CustomerID:         handler.ParseString(data, "customer_id"),
		PriceID:            handler.ParseString(data, "price_id"),
		Status:             status,
		AutoRenew:          autoRenew,
		StartAt:            startAt,
		CurrentPeriodStart: currentStart,
		CurrentPeriodEnd:   currentEnd,
		TrialStartAt:       trialStart,
		TrialEndAt:         trialEnd,
		CancelAt:           cancelAt,
		Metadata:           handler.MetadataMap(data),
	}, nil
}

func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
