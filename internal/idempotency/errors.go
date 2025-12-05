package idempotency

import "errors"

var (
	ErrMissingKey       = errors.New("missing idempotency key")
	ErrAlreadyCompleted = errors.New("request already completed")
)
