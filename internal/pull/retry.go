package pull

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Retry runs fn with exponential backoff until it succeeds, the context ends, or retryable reports false.
func Retry(ctx context.Context, boff backoff.BackOff, retryable func(error) bool, onRetry func(error, time.Duration), fn func() error) error {
	boff.Reset()
	op := func() error {
		err := fn()
		if err != nil && !retryable(err) {
			return backoff.Permanent(err)
		}
		return err
	}
	return backoff.RetryNotify(op, backoff.WithContext(boff, ctx), onRetry)
}

// Sleep waits for d or until the context ends, reporting whether the full duration elapsed.
func Sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}
