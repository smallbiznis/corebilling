package invoice

import (
	"context"
	"errors"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/handler"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	invoicedomain "github.com/smallbiznis/corebilling/internal/invoice/domain"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// InvoiceGeneratedHandler handles invoice.generated events.
type InvoiceGeneratedHandler struct {
	svc       *invoicedomain.Service
	tracker   *outbox.IdempotencyTracker
	publisher events.Publisher
	logger    *zap.Logger
}

// NewInvoiceGeneratedHandler constructs the handler.
func NewInvoiceGeneratedHandler(
	svc *invoicedomain.Service,
	publisher events.Publisher,
	tracker *outbox.IdempotencyTracker,
	logger *zap.Logger,
) handler.HandlerOut {
	return handler.HandlerOut{
		Handler: &InvoiceGeneratedHandler{
			svc:       svc,
			tracker:   tracker,
			publisher: publisher,
			logger:    logger.Named("invoice.generated"),
		},
	}
}

func (h *InvoiceGeneratedHandler) Subject() string {
	return "invoice.generated"
}

func (h *InvoiceGeneratedHandler) Handle(ctx context.Context, evt *events.Event) error {
	if evt == nil {
		return errors.New("event required")
	}
	if h.tracker != nil && h.tracker.SeenBefore(evt.GetId()) {
		h.logger.Debug("event already processed", zap.String("event_id", evt.GetId()))
		return nil
	}

	invoice, err := h.buildInvoice(evt)
	if err != nil {
		return err
	}
	if err := h.svc.Create(ctx, invoice); err != nil {
		return err
	}

	if h.publisher != nil {
		payload := map[string]*structpb.Value{
			"invoice_id":  structpb.NewStringValue(invoice.ID),
			"total_cents": structpb.NewNumberValue(float64(invoice.TotalCents)),
		}
		if child, childErr := handler.NewFollowUpEvent(evt, "invoice.sent", invoice.TenantID, payload); childErr == nil {
			_ = h.publisher.Publish(ctx, events.EventEnvelope{Event: child})
		}
	}
	return nil
}

func (h *InvoiceGeneratedHandler) buildInvoice(evt *events.Event) (invoicedomain.Invoice, error) {
	data := evt.GetData()
	issuedAt, err := handler.ParseTime(data, "issued_at")
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	dueAt, err := handler.ParseTime(data, "due_at")
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	metadata := handler.MetadataMap(handler.StructValue(data, "metadata"))
	return invoicedomain.Invoice{
		ID:             handler.ParseString(data, "invoice_id"),
		TenantID:       evt.GetTenantId(),
		CustomerID:     handler.ParseString(data, "customer_id"),
		SubscriptionID: handler.ParseString(data, "subscription_id"),
		Status:         int32(handler.ParseFloat(data, "status")),
		CurrencyCode:   handler.ParseString(data, "currency"),
		TotalCents:     int64(handler.ParseFloat(data, "total_cents")),
		SubtotalCents:  int64(handler.ParseFloat(data, "subtotal_cents")),
		TaxCents:       int64(handler.ParseFloat(data, "tax_cents")),
		InvoiceNumber:  handler.ParseString(data, "invoice_number"),
		IssuedAt:       issuedAt,
		DueAt:          dueAt,
		Metadata:       metadata,
	}, nil
}
