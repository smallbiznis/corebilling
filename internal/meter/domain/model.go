package domain

import "time"

// Meter represents a usage meter configuration.
type Meter struct {
	ID          string
	TenantID    string
	Code        string
	Aggregation int32
	Transform   int32
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
