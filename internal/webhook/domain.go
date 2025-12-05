package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Subscription represents a tenant webhook subscription.
type Subscription struct {
	ID         string
	TenantID   string
	TargetURL  string
	Secret     string
	EventTypes []string
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// DeliveryAttempt captures a delivery attempt for a webhook event.
type DeliveryAttempt struct {
	ID        int64
	WebhookID string
	EventID   string
	TenantID  string
	Payload   []byte
	Status    string
	AttemptNo int32
	NextRunAt time.Time
	LastError string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WebhookEvent carries the serialized event payload to deliver.
type WebhookEvent struct {
	ID       string
	Subject  string
	TenantID string
	Payload  []byte
}

// SignPayload generates HMAC-SHA256 signature header for the payload.
func SignPayload(secret string, body []byte) string {
	if secret == "" || len(body) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return fmt.Sprintf("sbwh_sig=v1:%s", hex.EncodeToString(mac.Sum(nil)))
}
