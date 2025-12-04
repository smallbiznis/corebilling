package outbox

import (
	"time"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// OutboxStatus represents the lifecycle state of an outbox event stored within metadata.
type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusDispatched OutboxStatus = "dispatched"
	OutboxStatusFailed     OutboxStatus = "failed"
	OutboxStatusDeadLetter OutboxStatus = "dead_letter"
)

const (
	metadataStatusKey      = "outbox_status"
	metadataRetryKey       = "outbox_retry_count"
	metadataNextAttemptKey = "outbox_next_attempt_at"
	metadataLastErrorKey   = "outbox_last_error"
)

// OutboxEvent represents a row in the billing_events table enriched with metadata-driven state.
type OutboxEvent struct {
	ID            string
	Subject       string
	TenantID      string
	ResourceID    string
	Payload       []byte
	Status        OutboxStatus
	RetryCount    int32
	NextAttemptAt time.Time
	LastError     string
	CreatedAt     time.Time
	Event         *eventv1.Event
}

// ExtractMetadata derives outbox metadata from the event payload.
func ExtractMetadata(evt *eventv1.Event) (OutboxStatus, int32, time.Time, string) {
	if evt == nil {
		return OutboxStatusPending, 0, time.Time{}, ""
	}

	m := evt.GetMetadata()
	if m == nil {
		return OutboxStatusPending, 0, time.Time{}, ""
	}

	data := m.AsMap()
	status := OutboxStatusPending
	if raw, ok := data[metadataStatusKey]; ok {
		if s, ok := raw.(string); ok && s != "" {
			status = OutboxStatus(s)
		}
	}

	var retry int32
	if raw, ok := data[metadataRetryKey]; ok {
		switch v := raw.(type) {
		case float64:
			retry = int32(v)
		case int64:
			retry = int32(v)
		case int:
			retry = int32(v)
		}
	}

	var nextAttempt time.Time
	if raw, ok := data[metadataNextAttemptKey]; ok {
		if s, ok := raw.(string); ok && s != "" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				nextAttempt = t
			}
		}
	}

	var lastErr string
	if raw, ok := data[metadataLastErrorKey]; ok {
		if s, ok := raw.(string); ok {
			lastErr = s
		}
	}

	return status, retry, nextAttempt, lastErr
}

// ApplyMetadata merges outbox metadata into the event metadata struct.
func ApplyMetadata(evt *eventv1.Event, status OutboxStatus, retryCount int32, nextAttemptAt time.Time, lastError string) (*eventv1.Event, error) {
	if evt.Metadata == nil {
		evt.Metadata = &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}

	setString(evt, metadataStatusKey, string(status))
	setNumber(evt, metadataRetryKey, float64(retryCount))
	if !nextAttemptAt.IsZero() {
		setString(evt, metadataNextAttemptKey, nextAttemptAt.UTC().Format(time.RFC3339))
	}
	if lastError != "" {
		setString(evt, metadataLastErrorKey, lastError)
	}
	return evt, nil
}

func setString(evt *eventv1.Event, key, val string) {
	if evt.Metadata.Fields == nil {
		evt.Metadata.Fields = map[string]*structpb.Value{}
	}
	evt.Metadata.Fields[key] = structpb.NewStringValue(val)
}

func setNumber(evt *eventv1.Event, key string, val float64) {
	if evt.Metadata.Fields == nil {
		evt.Metadata.Fields = map[string]*structpb.Value{}
	}
	evt.Metadata.Fields[key] = structpb.NewNumberValue(val)
}
