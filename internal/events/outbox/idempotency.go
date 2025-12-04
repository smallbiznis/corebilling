package outbox

import (
	"sync"
	"time"
)

// IdempotencyTracker keeps a short-lived cache of processed event ids.
type IdempotencyTracker struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	ttl     time.Duration
	stopCh  chan struct{}
	started bool
}

// NewIdempotencyTracker constructs a tracker with the given ttl.
func NewIdempotencyTracker(ttl time.Duration) *IdempotencyTracker {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &IdempotencyTracker{
		seen:   make(map[string]time.Time),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
}

// Start begins background cleanup.
func (t *IdempotencyTracker) Start() {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return
	}
	t.started = true
	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(t.ttl)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.cleanup()
			case <-t.stopCh:
				return
			}
		}
	}()
}

// Stop terminates cleanup goroutine.
func (t *IdempotencyTracker) Stop() {
	t.mu.Lock()
	if !t.started {
		t.mu.Unlock()
		return
	}
	t.started = false
	close(t.stopCh)
	t.mu.Unlock()
}

// SeenBefore returns true if the key has already been recorded within the TTL.
func (t *IdempotencyTracker) SeenBefore(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	expiration, ok := t.seen[key]
	if !ok {
		t.seen[key] = time.Now().Add(t.ttl)
		return false
	}
	if time.Now().After(expiration) {
		delete(t.seen, key)
		t.seen[key] = time.Now().Add(t.ttl)
		return false
	}
	return true
}

func (t *IdempotencyTracker) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for k, exp := range t.seen {
		if now.After(exp) {
			delete(t.seen, k)
		}
	}
}
