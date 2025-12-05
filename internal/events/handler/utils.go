package handler

import (
	"errors"
	"strings"
	"time"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// timeLayout is used for parsing RFC3339 timestamps.
const timeLayout = time.RFC3339

// ParseString returns the string value for a field.
func ParseString(data *structpb.Struct, key string) string {
	if data == nil || data.Fields == nil {
		return ""
	}
	if val, ok := data.Fields[key]; ok {
		return val.GetStringValue()
	}
	return ""
}

// ParseBool returns the boolean value for a field.
func ParseBool(data *structpb.Struct, key string) bool {
	if data == nil || data.Fields == nil {
		return false
	}
	if val, ok := data.Fields[key]; ok {
		return val.GetBoolValue()
	}
	return false
}

// ParseFloat returns the float value for a field.
func ParseFloat(data *structpb.Struct, key string) float64 {
	if data == nil || data.Fields == nil {
		return 0
	}
	if val, ok := data.Fields[key]; ok {
		return val.GetNumberValue()
	}
	return 0
}

// ParseTime parses a time field if present.
func ParseTime(data *structpb.Struct, key string) (*time.Time, error) {
	if data == nil || data.Fields == nil {
		return nil, nil
	}
	if val, ok := data.Fields[key]; ok {
		if s := strings.TrimSpace(val.GetStringValue()); s != "" {
			if t, err := time.Parse(timeLayout, s); err == nil {
				return &t, nil
			} else {
				return nil, err
			}
		}
	}
	return nil, nil
}

// MetadataMap returns the metadata struct map safely.
func MetadataMap(data *structpb.Struct) map[string]interface{} {
	if data == nil {
		return nil
	}
	return data.AsMap()
}

// EventData builds a structpb.Struct from a value map.
func EventData(values map[string]*structpb.Value) *structpb.Struct {
	if len(values) == 0 {
		return nil
	}
	return &structpb.Struct{Fields: values}
}

// NewFollowUpEvent creates a new child event linked to the parent.
func NewFollowUpEvent(parent *eventv1.Event, subject string, tenantID string, data map[string]*structpb.Value) (*eventv1.Event, error) {
	if subject == "" {
		return nil, errors.New("subject required")
	}

	evt := &eventv1.Event{
		Subject:   subject,
		TenantId:  tenantID,
		CreatedAt: timestamppb.Now(),
		Data:      EventData(data),
	}
	if parent != nil {
		ensureMetadata(evt)
		if parent.Metadata != nil && parent.Metadata.Fields != nil {
			if corr, ok := parent.Metadata.Fields["correlation_id"]; ok {
				evt.Metadata.Fields["correlation_id"] = corr
			}
		}
		evt.Metadata.Fields["causation_id"] = structpb.NewStringValue(parent.GetId())
	}
	return evt, nil
}

func ensureMetadata(evt *eventv1.Event) {
	if evt.Metadata == nil {
		evt.Metadata = &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	if evt.Metadata.Fields == nil {
		evt.Metadata.Fields = map[string]*structpb.Value{}
	}
}

// StructValue returns nested struct field.
func StructValue(data *structpb.Struct, key string) *structpb.Struct {
	if data == nil || data.Fields == nil {
		return nil
	}
	if val, ok := data.Fields[key]; ok {
		return val.GetStructValue()
	}
	return nil
}
