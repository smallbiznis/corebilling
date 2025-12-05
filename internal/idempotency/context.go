package idempotency

import "context"

type contextKey struct{}

// WithIdempotencyKey stores the idempotency key in context.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, contextKey{}, key)
}

// ExtractIdempotencyKey retrieves the idempotency key from context.
func ExtractIdempotencyKey(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(contextKey{}).(string); ok {
		return val
	}
	return ""
}
