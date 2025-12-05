package idempotency

import "time"

type Status string

const (
	StatusProcessing Status = "PROCESSING"
	StatusCompleted  Status = "COMPLETED"
)

type Record struct {
	TenantID    string
	Key         string
	RequestHash string
	Response    []byte
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
