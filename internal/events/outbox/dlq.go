package outbox

const MaxRetries = 10

// ShouldMoveToDLQ determines if retry budget is exhausted.
func ShouldMoveToDLQ(retryCount int32) bool {
	return retryCount >= MaxRetries
}
