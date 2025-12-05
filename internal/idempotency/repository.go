package idempotency

import "context"

type Repository interface {
	Get(ctx context.Context, tenantID, key string) (*Record, error)
	InsertProcessing(ctx context.Context, tenantID, key, requestHash string) error
	MarkCompleted(ctx context.Context, tenantID, key string, response []byte) error
}
