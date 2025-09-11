package structs

import (
	"context"
	Error "sentinel/packages/common/errors"
	"time"
)

// Returns Error.StatusTimeout on timeout.
func SetTimeout(ctx context.Context, timeout time.Duration, req func(ctx context.Context)) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		req(ctx)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		err := ctx.Err()
		if err == context.DeadlineExceeded {
			return Error.StatusTimeout
		}
		return err
	}
}

