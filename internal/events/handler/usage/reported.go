package usage

import (
	"context"
	"errors"
	"time"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/handler"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	usagedomain "github.com/smallbiznis/corebilling/internal/usage/domain"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// UsageReportedHandler processes usage.reported events.
type UsageReportedHandler struct {
	svc       *usagedomain.Service
	tracker   *outbox.IdempotencyTracker
	publisher events.Publisher
	logger    *zap.Logger
}

// NewUsageReportedHandler constructs the handler.
func NewUsageReportedHandler(
	svc *usagedomain.Service,
	publisher events.Publisher,
	tracker *outbox.IdempotencyTracker,
	logger *zap.Logger,
) handler.HandlerOut {
	return handler.HandlerOut{
		Handler: &UsageReportedHandler{
			svc:       svc,
			tracker:   tracker,
			publisher: publisher,
			logger:    logger.Named("usage.reported"),
		},
	}
}

func (h *UsageReportedHandler) Subject() string {
	return "usage.reported"
}

func (h *UsageReportedHandler) Handle(ctx context.Context, evt *events.Event) error {
	if evt == nil {
		return errors.New("event required")
	}
	if h.tracker != nil && h.tracker.SeenBefore(evt.GetId()) {
		h.logger.Debug("event already processed", zap.String("event_id", evt.GetId()))
		return nil
	}

	record, err := h.buildUsage(evt)
	if err != nil {
		return err
	}
	if err := h.svc.Create(ctx, record); err != nil {
		return err
	}

	if h.publisher != nil {
		payload := map[string]*structpb.Value{
			"usage_id":        structpb.NewStringValue(record.ID),
			"subscription_id": structpb.NewStringValue(record.SubscriptionID),
			"value":           structpb.NewNumberValue(record.Value),
		}
		if child, childErr := handler.NewFollowUpEvent(evt, "usage.rated", record.TenantID, payload); childErr == nil {
			_ = h.publisher.Publish(ctx, events.EventEnvelope{Event: child})
		}
	}
	return nil
}

func (h *UsageReportedHandler) buildUsage(evt *events.Event) (usagedomain.UsageRecord, error) {
	data := evt.GetData()
	recordedAt, err := handler.ParseTime(data, "recorded_at")
	if err != nil {
		return usagedomain.UsageRecord{}, err
	}
	return usagedomain.UsageRecord{
		ID:             handler.ParseString(data, "id"),
		TenantID:       evt.GetTenantId(),
		CustomerID:     handler.ParseString(data, "customer_id"),
		SubscriptionID: handler.ParseString(data, "subscription_id"),
		MeterCode:      handler.ParseString(data, "meter_code"),
		Value:          handler.ParseFloat(data, "value"),
		RecordedAt:     timeOrNow(recordedAt),
		IdempotencyKey: handler.ParseString(data, "idempotency_key"),
		Metadata:       handler.MetadataMap(handler.StructValue(data, "metadata")),
	}, nil
}

func timeOrNow(t *time.Time) time.Time {
	if t == nil {
		return time.Now().UTC()
	}
	return *t
}
