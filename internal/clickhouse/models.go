package clickhouse

import "time"

// UsageEvent represents a usage record destined for ClickHouse analytics.
type UsageEvent struct {
	TenantID       string
	CustomerID     string
	SubscriptionID string
	MeterCode      string
	Value          float64
	RecordedAt     time.Time
	IdempotencyKey string
	Metadata       string
}

// BillingEventLog models the billing event log entry for ClickHouse.
type BillingEventLog struct {
	EventID   string
	TenantID  string
	EventType string
	Payload   string
	CreatedAt time.Time
}
