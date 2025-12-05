package events

import (
	"context"
	"errors"

	"github.com/oklog/ulid/v2"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EventEnvelope carries the canonical event payload plus delivery metadata.
type EventEnvelope struct {
	Event         *eventv1.Event
	Subject       string
	TenantID      string
	CorrelationID string
	CausationID   string
	Payload       []byte
}

// NewEnvelope creates an envelope around an existing event, copying cached metadata.
func NewEnvelope(evt *eventv1.Event) EventEnvelope {
	env := EventEnvelope{Event: evt}
	if evt == nil {
		return env
	}
	env.Subject = evt.Subject
	env.TenantID = evt.TenantId
	if fields := evt.Metadata; fields != nil && fields.Fields != nil {
		env.CorrelationID = getMetadataString(fields.Fields, "correlation_id")
		env.CausationID = getMetadataString(fields.Fields, "causation_id")
	}
	return env
}

// Prepare ensures the envelope has an event, metadata, timestamps, and serialized payload.
func (e *EventEnvelope) Prepare(ctx context.Context) (context.Context, error) {
	evt, err := e.bindEvent()
	if err != nil {
		return ctx, err
	}
	if evt.Id == "" {
		evt.Id = ulid.Make().String()
	}
	if e.Subject == "" {
		e.Subject = evt.Subject
	}
	if evt.Subject == "" {
		evt.Subject = e.Subject
	}
	if e.TenantID == "" {
		e.TenantID = evt.TenantId
	}
	if evt.TenantId == "" {
		evt.TenantId = e.TenantID
	}
	if evt.CreatedAt == nil {
		evt.CreatedAt = timestamppb.Now()
	}
	ensureMetadata(evt)

	ctx, cid := correlation.EnsureCorrelationID(ctx)
	if e.CorrelationID == "" {
		e.CorrelationID = cid
	} else {
		ctx = correlation.ContextWithCorrelationID(ctx, e.CorrelationID)
	}
	setMetadataField(evt, "correlation_id", e.CorrelationID)

	if e.CausationID == "" {
		if val := getMetadataString(evt.Metadata.Fields, "causation_id"); val != "" {
			e.CausationID = val
		}
	}
	if e.CausationID != "" {
		setMetadataField(evt, "causation_id", e.CausationID)
	}

	if e.TenantID != "" {
		setMetadataField(evt, "tenant_id", e.TenantID)
	}
	if e.Subject != "" {
		setMetadataField(evt, "subject", e.Subject)
	}

	payload, err := MarshalEvent(evt)
	if err != nil {
		return ctx, err
	}
	e.Payload = payload
	return ctx, nil
}

func (e *EventEnvelope) bindEvent() (*eventv1.Event, error) {
	if e.Event != nil {
		return e.Event, nil
	}
	if len(e.Payload) == 0 {
		return nil, errors.New("event payload required")
	}
	evt, err := UnmarshalEvent(e.Payload)
	if err != nil {
		return nil, err
	}
	e.Event = evt
	return evt, nil
}

func ensureMetadata(evt *eventv1.Event) {
	if evt.Metadata == nil {
		evt.Metadata = &structpb.Struct{Fields: map[string]*structpb.Value{}}
		return
	}
	if evt.Metadata.Fields == nil {
		evt.Metadata.Fields = map[string]*structpb.Value{}
	}
}

func setMetadataField(evt *eventv1.Event, key, value string) {
	if value == "" {
		return
	}
	ensureMetadata(evt)
	evt.Metadata.Fields[key] = structpb.NewStringValue(value)
}

func getMetadataString(fields map[string]*structpb.Value, key string) string {
	if fields == nil {
		return ""
	}
	if val, ok := fields[key]; ok {
		return val.GetStringValue()
	}
	return ""
}
