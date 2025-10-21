package structs

import (
	"context"
	Error "sentinel/packages/common/errors"
	"time"
)

// Returns Error.StatusTimeout on timeout.
func SetTimeout(ctx context.Context, timeout time.Duration, req func(ctx context.Context)) error {
	// If timeout is zero or negative, don't set a timeout
	if timeout <= 0 {
		done := make(chan struct{})
		go func() {
			req(ctx)
			close(done)
		}()
		<-done
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		req(ctx)
		close(done)
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
