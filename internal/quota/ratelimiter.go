package quota

import (
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(tenantID string) bool
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucketRateLimiter struct {
	rate    float64
	burst   float64
	buckets map[string]*tokenBucket
	mu      sync.Mutex
}

// NewRateLimiter constructs a token bucket limiter with the provided rate and burst sizes.
func NewRateLimiter(rate float64, burst int) RateLimiter {
	if rate <= 0 {
		rate = 10
	}
	if burst <= 0 {
		burst = 100
	}
	return &TokenBucketRateLimiter{
		rate:    rate,
		burst:   float64(burst),
		buckets: make(map[string]*tokenBucket),
	}
}

// Allow attempts to consume one token for the tenant and reports if the request is permitted.
func (r *TokenBucketRateLimiter) Allow(tenantID string) bool {
	if tenantID == "" {
		return true
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, ok := r.buckets[tenantID]
	now := time.Now()
	if !ok {
		bucket = &tokenBucket{tokens: r.burst, lastRefill: now}
		r.buckets[tenantID] = bucket
	}

	elapsed := now.Sub(bucket.lastRefill).Seconds()
	if elapsed > 0 {
		bucket.tokens += elapsed * r.rate
		if bucket.tokens > r.burst {
			bucket.tokens = r.burst
		}
		bucket.lastRefill = now
	}

	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens -= 1
	return true
}
