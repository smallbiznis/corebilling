package e2e

import (
	"context"
	"time"
)

// waitUntil polls fn until it returns true or timeout.
func waitUntil(ctx context.Context, interval time.Duration, fn func(context.Context) (bool, error)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		ok, err := fn(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
