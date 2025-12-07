package clickhouse

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/handler"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Writer consumes bus events and persists analytics copies into ClickHouse.
type Writer struct {
	bus           events.Bus
	client        Client
	cfg           Config
	logger        *zap.Logger
	usageCh       chan *eventv1.Event
	billingCh     chan *eventv1.Event
	cancel        context.CancelFunc
	flushAttempts int
}

// NewWriter builds a ClickHouse writer.
func NewWriter(bus events.Bus, client Client, cfg Config, logger *zap.Logger) *Writer {
	return &Writer{
		bus:           bus,
		client:        client,
		cfg:           cfg,
		logger:        logger.Named("clickhouse.writer"),
		usageCh:       make(chan *eventv1.Event, cfg.BatchSize*2),
		billingCh:     make(chan *eventv1.Event, cfg.BatchSize*2),
		flushAttempts: 5,
	}
}

// Run wires lifecycle hooks for the writer.
func Run(lc fx.Lifecycle, w *Writer, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithCancel(ctx)
			w.cancel = cancel
			if err := w.subscribe(ctx); err != nil {
				return err
			}
			go w.loop(ctx)
			logger.Info("clickhouse writer started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if w.cancel != nil {
				w.cancel()
			}
			w.flushPending(ctx)
			logger.Info("clickhouse writer stopped")
			return nil
		},
	})
}

func (w *Writer) subscribe(ctx context.Context) error {
	if err := w.bus.Subscribe(ctx, "usage.recorded", "clickhouse-writer", w.handleUsage); err != nil {
		return err
	}
	if err := w.bus.Subscribe(ctx, "billing.event.logged", "clickhouse-writer", w.handleBillingEvent); err != nil {
		return err
	}
	return nil
}

func (w *Writer) handleUsage(ctx context.Context, evt *eventv1.Event) error {
	select {
	case w.usageCh <- evt:
	default:
		w.logger.Warn("usage channel full, dropping event", zap.String("tenant_id", evt.GetTenantId()), zap.String("event_id", evt.GetId()))
	}
	return nil
}

func (w *Writer) handleBillingEvent(ctx context.Context, evt *eventv1.Event) error {
	select {
	case w.billingCh <- evt:
	default:
		w.logger.Warn("billing channel full, dropping event", zap.String("tenant_id", evt.GetTenantId()), zap.String("event_id", evt.GetId()))
	}
	return nil
}

func (w *Writer) loop(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.FlushInterval)
	defer ticker.Stop()

	usageBatch := make([]UsageEvent, 0, w.cfg.BatchSize)
	billingBatch := make([]BillingEventLog, 0, w.cfg.BatchSize)

	flushUsage := func() {
		if len(usageBatch) == 0 {
			return
		}
		batch := make([]UsageEvent, len(usageBatch))
		copy(batch, usageBatch)
		usageBatch = usageBatch[:0]
		w.flushUsage(ctx, batch)
	}
	flushBilling := func() {
		if len(billingBatch) == 0 {
			return
		}
		batch := make([]BillingEventLog, len(billingBatch))
		copy(batch, billingBatch)
		billingBatch = billingBatch[:0]
		w.flushBilling(ctx, batch)
	}

	for {
		select {
		case <-ctx.Done():
			flushUsage()
			flushBilling()
			return
		case evt := <-w.usageCh:
			if evt != nil {
				if record, err := w.mapUsage(evt); err == nil {
					usageBatch = append(usageBatch, record)
					if len(usageBatch) >= w.cfg.BatchSize {
						flushUsage()
					}
				} else {
					w.logger.Warn("failed to map usage event", zap.Error(err), zap.String("tenant_id", evt.GetTenantId()))
				}
			}
		case evt := <-w.billingCh:
			if evt != nil {
				if record, err := w.mapBilling(evt); err == nil {
					billingBatch = append(billingBatch, record)
					if len(billingBatch) >= w.cfg.BatchSize {
						flushBilling()
					}
				} else {
					w.logger.Warn("failed to map billing event", zap.Error(err), zap.String("tenant_id", evt.GetTenantId()))
				}
			}
		case <-ticker.C:
			flushUsage()
			flushBilling()
		}
	}
}

func (w *Writer) flushUsage(ctx context.Context, batch []UsageEvent) {
	err := retryWithBackoff(ctx, w.flushAttempts, 200*time.Millisecond, func() error {
		return w.client.InsertUsageBatch(ctx, batch)
	})
	if err != nil {
		w.logger.Error("failed to write usage batch", zap.Error(err), zap.Int("batch_size", len(batch)), zap.String("tenant_ids", tenantSummaryUsage(batch)))
	}
}

func (w *Writer) flushBilling(ctx context.Context, batch []BillingEventLog) {
	err := retryWithBackoff(ctx, w.flushAttempts, 200*time.Millisecond, func() error {
		return w.client.InsertBillingEventBatch(ctx, batch)
	})
	if err != nil {
		w.logger.Error("failed to write billing batch", zap.Error(err), zap.Int("batch_size", len(batch)), zap.String("tenant_ids", tenantSummaryBilling(batch)))
	}
}

func (w *Writer) flushPending(ctx context.Context) {
	for {
		if len(w.usageCh) == 0 && len(w.billingCh) == 0 {
			return
		}
		select {
		case evt := <-w.usageCh:
			if evt != nil {
				if record, err := w.mapUsage(evt); err == nil {
					w.flushUsage(ctx, []UsageEvent{record})
				}
			}
		case evt := <-w.billingCh:
			if evt != nil {
				if record, err := w.mapBilling(evt); err == nil {
					w.flushBilling(ctx, []BillingEventLog{record})
				}
			}
		default:
			return
		}
	}
}

func (w *Writer) mapUsage(evt *eventv1.Event) (UsageEvent, error) {
	if evt == nil {
		return UsageEvent{}, nil
	}
	data := evt.GetData()
	recordedAt, err := handler.ParseTime(data, "recorded_at")
	if err != nil {
		return UsageEvent{}, err
	}
	metadata := handler.StructValue(data, "metadata")
	metadataJSON := ""
	if metadata != nil {
		if b, err := json.Marshal(metadata.AsMap()); err == nil {
			metadataJSON = string(b)
		}
	}

	return UsageEvent{
		TenantID:       evt.GetTenantId(),
		CustomerID:     handler.ParseString(data, "customer_id"),
		SubscriptionID: handler.ParseString(data, "subscription_id"),
		MeterCode:      handler.ParseString(data, "meter_code"),
		Value:          handler.ParseFloat(data, "value"),
		RecordedAt:     timeOrNowPtr(recordedAt),
		IdempotencyKey: handler.ParseString(data, "idempotency_key"),
		Metadata:       metadataJSON,
	}, nil
}

func (w *Writer) mapBilling(evt *eventv1.Event) (BillingEventLog, error) {
	if evt == nil {
		return BillingEventLog{}, nil
	}
	payload := ""
	if evt.GetData() != nil {
		if b, err := json.Marshal(evt.GetData().AsMap()); err == nil {
			payload = string(b)
		}
	}
	return BillingEventLog{
		EventID:   evt.GetId(),
		TenantID:  evt.GetTenantId(),
		EventType: evt.GetSubject(),
		Payload:   payload,
		CreatedAt: timeOrNowTime(evt.GetCreatedAt()),
	}, nil
}

func tenantSummaryUsage(batch []UsageEvent) string {
	tenants := make(map[string]struct{})
	for _, r := range batch {
		tenants[r.TenantID] = struct{}{}
	}
	parts := make([]string, 0, len(tenants))
	for t := range tenants {
		parts = append(parts, t)
	}
	return joinTenants(parts)
}

func tenantSummaryBilling(batch []BillingEventLog) string {
	tenants := make(map[string]struct{})
	for _, r := range batch {
		tenants[r.TenantID] = struct{}{}
	}
	parts := make([]string, 0, len(tenants))
	for t := range tenants {
		parts = append(parts, t)
	}
	return joinTenants(parts)
}

func joinTenants(tenants []string) string {
	if len(tenants) == 0 {
		return ""
	}
	return strings.Join(tenants, ",")
}

func timeOrNowPtr(t *time.Time) time.Time {
	if t == nil {
		return time.Now().UTC()
	}
	return *t
}

func timeOrNowTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Now().UTC()
	}
	return t.AsTime()
}
