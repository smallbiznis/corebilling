package headers

import (
	"context"
	"net/textproto"
	"reflect"

	"google.golang.org/grpc/metadata"
)

// HeaderValues captures exposed metadata.
type HeaderValues struct {
	TenantID       string
	APIKey         string
	IdempotencyKey string
	EventID        string
	EventType      string
	CorrelationID  string
	CausationID    string
}

// ExtractMetadata reads recognized keys from metadata.MD.
func ExtractMetadata(md metadata.MD) HeaderValues {
	if md == nil {
		return HeaderValues{}
	}
	return HeaderValues{
		TenantID:       first(md.Get(MetadataTenantID)),
		APIKey:         first(md.Get(MetadataAPIKey)),
		IdempotencyKey: first(md.Get(MetadataIdempotency)),
		EventID:        first(md.Get(MetadataEventID)),
		EventType:      first(md.Get(MetadataEventType)),
		CorrelationID:  first(md.Get(MetadataCorrelation)),
		CausationID:    first(md.Get(MetadataCausation)),
	}
}

// MergeIdempotencyKey prefers header over body.
func MergeIdempotencyKey(headerVal, bodyVal string) string {
	if headerVal != "" {
		return headerVal
	}
	return bodyVal
}

// ApplyHeadersToRequest copies tenant/idempotency headers into request struct fields.
func ApplyHeadersToRequest(ctx context.Context, req interface{}, md metadata.MD) {
	if req == nil || md == nil {
		return
	}
	val := reflect.ValueOf(req)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}
	val = val.Elem()
	tenant := first(md.Get(MetadataTenantID))
	if tenant != "" {
		setStringField(val, "TenantID", tenant)
	}
	if idVal := first(md.Get(MetadataIdempotency)); idVal != "" {
		setStringField(val, "IdempotencyKey", idVal)
	}
}

// NormalizeHeader returns canonical HTTP header casing.
func NormalizeHeader(h string) string {
	return textproto.CanonicalMIMEHeaderKey(h)
}
func first(vals []string) string {
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func setStringField(val reflect.Value, name, value string) {
	field := val.FieldByName(name)
	if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
		return
	}
	field.SetString(value)
}
