package outbox

import "time"

// ComputeNextAttempt calculates exponential backoff capped by max.
func ComputeNextAttempt(retryCount int32, base time.Duration, max time.Duration) time.Time {
	if base <= 0 {
		base = time.Second
	}
	delay := base << retryCount
	if delay > max && max > 0 {
		delay = max
	}
	return time.Now().Add(delay)
}
