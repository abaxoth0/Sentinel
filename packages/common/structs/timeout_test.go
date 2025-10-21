package structs

import (
	"context"
	"testing"
	"time"
)

func TestSetTimeout(t *testing.T) {
	t.Run("function completes before timeout", func(t *testing.T) {
		ctx := context.Background()
		timeout := 100 * time.Millisecond

		executed := false
		req := func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
			executed = true
		}

		err := SetTimeout(ctx, timeout, req)
		if err != nil {
			t.Errorf("SetTimeout should succeed: %v", err)
		}

		if !executed {
			t.Error("Function should have been executed")
		}
	})

	t.Run("function times out", func(t *testing.T) {
		ctx := context.Background()
		timeout := 50 * time.Millisecond

		req := func(ctx context.Context) {
			time.Sleep(100 * time.Millisecond)
		}

		err := SetTimeout(ctx, timeout, req)
		if err == nil {
			t.Error("SetTimeout should return timeout error")
		}

		// Function might still be running, but we don't care
		// The important thing is that SetTimeout returned with timeout error
	})

	t.Run("function respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		timeout := 100 * time.Millisecond

		executed := false
		req := func(ctx context.Context) {
			// Wait for context cancellation
			<-ctx.Done()
			executed = true
		}

		// Cancel context after a short delay
		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()

		err := SetTimeout(ctx, timeout, req)
		if err == nil {
			t.Error("SetTimeout should return context cancellation error")
		}

		// Give function time to execute
		time.Sleep(50 * time.Millisecond)

		if !executed {
			t.Error("Function should have been executed")
		}
	})

	t.Run("zero timeout", func(t *testing.T) {
		ctx := context.Background()
		timeout := 0 * time.Millisecond

		executed := false
		req := func(ctx context.Context) {
			executed = true
		}

		err := SetTimeout(ctx, timeout, req)
		// Zero timeout means no timeout, so function should execute
		if err != nil {
			t.Errorf("SetTimeout with zero timeout should succeed: %v", err)
		}
		if !executed {
			t.Error("Function should have been executed")
		}
	})

	t.Run("negative timeout", func(t *testing.T) {
		ctx := context.Background()
		timeout := -10 * time.Millisecond

		executed := false
		req := func(ctx context.Context) {
			executed = true
		}

		err := SetTimeout(ctx, timeout, req)
		// Negative timeout means no timeout, so function should execute
		if err != nil {
			t.Errorf("SetTimeout with negative timeout should succeed: %v", err)
		}
		if !executed {
			t.Error("Function should have been executed")
		}
	})

	t.Run("function completes quickly", func(t *testing.T) {
		ctx := context.Background()
		timeout := 100 * time.Millisecond

		executed := false
		req := func(ctx context.Context) {
			// Simulate quick execution
			time.Sleep(1 * time.Millisecond)
			executed = true
		}

		err := SetTimeout(ctx, timeout, req)
		if err != nil {
			t.Errorf("SetTimeout should succeed: %v", err)
		}

		if !executed {
			t.Error("Function should have been executed")
		}
	})

	t.Run("multiple concurrent timeouts", func(t *testing.T) {
		ctx := context.Background()
		timeout := 50 * time.Millisecond

		const numGoroutines = 10
		results := make([]bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				executed := false
				req := func(ctx context.Context) {
					time.Sleep(20 * time.Millisecond)
					executed = true
				}

				err := SetTimeout(ctx, timeout, req)
				results[index] = err == nil && executed
			}(i)
		}

		// Wait for all goroutines to complete
		time.Sleep(200 * time.Millisecond)

		// All should succeed
		for i, result := range results {
			if !result {
				t.Errorf("Goroutine %d should have succeeded", i)
			}
		}
	})
}
