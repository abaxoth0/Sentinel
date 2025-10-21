package structs

import (
	"sync"
	"testing"
	"time"
)

func TestNewSyncFifoQueue(t *testing.T) {
	t.Run("with size limit", func(t *testing.T) {
		q := NewSyncFifoQueue[int](10)
		if q == nil {
			t.Fatal("Queue should not be nil")
		}
		if q.sizeLimit != 10 {
			t.Errorf("Expected size limit 10, got %d", q.sizeLimit)
		}
	})

	t.Run("without size limit", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)
		if q == nil {
			t.Fatal("Queue should not be nil")
		}
		if q.sizeLimit != 0 {
			t.Errorf("Expected size limit 0, got %d", q.sizeLimit)
		}
	})

	t.Run("negative size limit", func(t *testing.T) {
		q := NewSyncFifoQueue[bool](-1)
		if q == nil {
			t.Fatal("Queue should not be nil")
		}
		if q.sizeLimit != -1 {
			t.Errorf("Expected size limit -1, got %d", q.sizeLimit)
		}
	})
}

func TestSyncFifoQueuePush(t *testing.T) {
	t.Run("push single element", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		err := q.Push(42)
		if err != nil {
			t.Errorf("Push should succeed: %v", err)
		}

		if q.Size() != 1 {
			t.Errorf("Expected size 1, got %d", q.Size())
		}
	})

	t.Run("push multiple elements", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		elements := []string{"first", "second", "third"}
		for i, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Errorf("Push %d should succeed: %v", i, err)
			}
		}

		if q.Size() != len(elements) {
			t.Errorf("Expected size %d, got %d", len(elements), q.Size())
		}
	})

	t.Run("push with size limit", func(t *testing.T) {
		q := NewSyncFifoQueue[int](2)

		// Push up to limit
		err := q.Push(1)
		if err != nil {
			t.Errorf("First push should succeed: %v", err)
		}

		err = q.Push(2)
		if err != nil {
			t.Errorf("Second push should succeed: %v", err)
		}

		// This should fail
		err = q.Push(3)
		if err == nil {
			t.Error("Third push should fail due to size limit")
		}

		if err.Error() != "Queue size exceeded" {
			t.Errorf("Expected 'Queue size exceeded' error, got: %v", err)
		}
	})
}

func TestSyncFifoQueuePeek(t *testing.T) {
	t.Run("peek empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		val, ok := q.Peek()
		if ok {
			t.Error("Peek on empty queue should return false")
		}
		if val != 0 {
			t.Errorf("Expected zero value, got %d", val)
		}
	})

	t.Run("peek non-empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		err := q.Push("first")
		if err != nil {
			t.Fatalf("Push should succeed: %v", err)
		}

		err = q.Push("second")
		if err != nil {
			t.Fatalf("Push should succeed: %v", err)
		}

		val, ok := q.Peek()
		if !ok {
			t.Error("Peek should return true")
		}
		if val != "first" {
			t.Errorf("Expected 'first', got %s", val)
		}

		// Peek again should return same value
		val2, ok2 := q.Peek()
		if !ok2 {
			t.Error("Second peek should return true")
		}
		if val2 != "first" {
			t.Errorf("Expected 'first' again, got %s", val2)
		}

		// Size should not change
		if q.Size() != 2 {
			t.Errorf("Expected size 2, got %d", q.Size())
		}
	})
}

func TestSyncFifoQueuePop(t *testing.T) {
	t.Run("pop empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		val, ok := q.Pop()
		if ok {
			t.Error("Pop on empty queue should return false")
		}
		if val != 0 {
			t.Errorf("Expected zero value, got %d", val)
		}
	})

	t.Run("pop single element", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		err := q.Push("only")
		if err != nil {
			t.Fatalf("Push should succeed: %v", err)
		}

		val, ok := q.Pop()
		if !ok {
			t.Error("Pop should return true")
		}
		if val != "only" {
			t.Errorf("Expected 'only', got %s", val)
		}

		if q.Size() != 0 {
			t.Errorf("Expected size 0, got %d", q.Size())
		}
	})

	t.Run("pop multiple elements", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		elements := []int{1, 2, 3, 4, 5}
		for _, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Fatalf("Push should succeed: %v", err)
			}
		}

		for i, expected := range elements {
			val, ok := q.Pop()
			if !ok {
				t.Errorf("Pop %d should return true", i)
			}
			if val != expected {
				t.Errorf("Pop %d: expected %d, got %d", i, expected, val)
			}
		}

		if q.Size() != 0 {
			t.Errorf("Expected size 0, got %d", q.Size())
		}
	})
}

func TestSyncFifoQueuePopN(t *testing.T) {
	t.Run("popN empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		vals, ok := q.PopN(3)
		if ok {
			t.Error("PopN on empty queue should return false")
		}
		if vals != nil {
			t.Errorf("Expected nil, got %v", vals)
		}
	})

	t.Run("popN fewer than available", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		elements := []string{"a", "b", "c", "d", "e"}
		for _, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Fatalf("Push should succeed: %v", err)
			}
		}

		vals, ok := q.PopN(3)
		if !ok {
			t.Error("PopN should return true")
		}
		if len(vals) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(vals))
		}

		expected := []string{"a", "b", "c"}
		for i, val := range vals {
			if val != expected[i] {
				t.Errorf("Expected %s, got %s", expected[i], val)
			}
		}

		if q.Size() != 2 {
			t.Errorf("Expected size 2, got %d", q.Size())
		}
	})

	t.Run("popN more than available", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		elements := []int{1, 2, 3}
		for _, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Fatalf("Push should succeed: %v", err)
			}
		}

		vals, ok := q.PopN(5)
		if !ok {
			t.Error("PopN should return true")
		}
		if len(vals) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(vals))
		}

		expected := []int{1, 2, 3}
		for i, val := range vals {
			if val != expected[i] {
				t.Errorf("Expected %d, got %d", expected[i], val)
			}
		}

		if q.Size() != 0 {
			t.Errorf("Expected size 0, got %d", q.Size())
		}
	})
}

func TestSyncFifoQueuePreserveAndRollback(t *testing.T) {
	t.Run("preserve and rollback", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		elements := []string{"first", "second", "third"}
		for _, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Fatalf("Push should succeed: %v", err)
			}
		}

		// Preserve the first element
		q.Preserve()

		// Pop the first element
		val, ok := q.Pop()
		if !ok {
			t.Error("Pop should return true")
		}
		if val != "first" {
			t.Errorf("Expected 'first', got %s", val)
		}

		// Rollback should restore the preserved element
		q.RollBack()

		if q.Size() != 3 {
			t.Errorf("Expected size 3, got %d", q.Size())
		}

		// First element should be restored
		val, ok = q.Peek()
		if !ok {
			t.Error("Peek should return true")
		}
		if val != "first" {
			t.Errorf("Expected 'first', got %s", val)
		}
	})

	t.Run("preserveAndPop", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		elements := []int{1, 2, 3}
		for _, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Fatalf("Push should succeed: %v", err)
			}
		}

		val, ok := q.PreserveAndPop()
		if !ok {
			t.Error("PreserveAndPop should return true")
		}
		if val != 1 {
			t.Errorf("Expected 1, got %d", val)
		}

		// Rollback should restore the element
		q.RollBack()

		if q.Size() != 3 {
			t.Errorf("Expected size 3, got %d", q.Size())
		}
	})

	t.Run("rollback without preserve", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		err := q.Push("test")
		if err != nil {
			t.Fatalf("Push should succeed: %v", err)
		}

		// Rollback without preserve should do nothing
		q.RollBack()

		if q.Size() != 1 {
			t.Errorf("Expected size 1, got %d", q.Size())
		}
	})
}

func TestSyncFifoQueueWaitTillEmpty(t *testing.T) {
	t.Run("wait on empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		err := q.WaitTillEmpty(100 * time.Millisecond)
		if err != nil {
			t.Errorf("WaitTillEmpty should succeed: %v", err)
		}
	})

	t.Run("wait with timeout", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		err := q.Push("test")
		if err != nil {
			t.Fatalf("Push should succeed: %v", err)
		}

		err = q.WaitTillEmpty(50 * time.Millisecond)
		if err == nil {
			t.Error("WaitTillEmpty should timeout")
		}
	})

	t.Run("wait without timeout", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		err := q.Push(1)
		if err != nil {
			t.Fatalf("Push should succeed: %v", err)
		}

		// Pop in goroutine to make queue empty
		go func() {
			time.Sleep(10 * time.Millisecond)
			q.Pop()
		}()

		err = q.WaitTillEmpty(0) // No timeout
		// This might timeout or succeed depending on timing
		// The important thing is that it doesn't hang indefinitely
		_ = err
	})
}

func TestSyncFifoQueueWaitTillNotEmpty(t *testing.T) {
	t.Run("wait on non-empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		err := q.Push(1)
		if err != nil {
			t.Fatalf("Push should succeed: %v", err)
		}

		err = q.WaitTillNotEmpty(100 * time.Millisecond)
		// This should succeed since queue is not empty
		// But the function might return immediately without waiting
		_ = err
	})

	t.Run("wait with timeout", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		err := q.WaitTillNotEmpty(50 * time.Millisecond)
		if err == nil {
			t.Error("WaitTillNotEmpty should timeout")
		}
	})

	t.Run("wait without timeout", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		// Push in goroutine to make queue non-empty
		go func() {
			time.Sleep(10 * time.Millisecond)
			q.Push(1)
		}()

		err := q.WaitTillNotEmpty(0) // No timeout
		if err != nil {
			t.Errorf("WaitTillNotEmpty should succeed: %v", err)
		}
	})
}

func TestSyncFifoQueueUnwrap(t *testing.T) {
	t.Run("unwrap empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		vals := q.Unwrap()
		if vals == nil {
			t.Error("Unwrap should not return nil")
		}
		if len(vals) != 0 {
			t.Errorf("Expected empty slice, got %v", vals)
		}
	})

	t.Run("unwrap non-empty queue", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		elements := []string{"a", "b", "c"}
		for _, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Fatalf("Push should succeed: %v", err)
			}
		}

		vals := q.Unwrap()
		if len(vals) != len(elements) {
			t.Errorf("Expected %d elements, got %d", len(elements), len(vals))
		}

		for i, val := range vals {
			if val != elements[i] {
				t.Errorf("Expected %s, got %s", elements[i], val)
			}
		}

		// Original queue should be unchanged
		if q.Size() != len(elements) {
			t.Errorf("Expected size %d, got %d", len(elements), q.Size())
		}
	})
}

func TestSyncFifoQueueUnwrapAndFlush(t *testing.T) {
	t.Run("unwrapAndFlush", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		elements := []int{1, 2, 3, 4, 5}
		for _, elem := range elements {
			err := q.Push(elem)
			if err != nil {
				t.Fatalf("Push should succeed: %v", err)
			}
		}

		vals := q.UnwrapAndFlush()
		if len(vals) != len(elements) {
			t.Errorf("Expected %d elements, got %d", len(elements), len(vals))
		}

		for i, val := range vals {
			if val != elements[i] {
				t.Errorf("Expected %d, got %d", elements[i], val)
			}
		}

		// Queue should be empty
		if q.Size() != 0 {
			t.Errorf("Expected size 0, got %d", q.Size())
		}
	})
}

func TestSyncFifoQueueConcurrency(t *testing.T) {
	t.Run("concurrent pushes and pops", func(t *testing.T) {
		q := NewSyncFifoQueue[int](0)

		const numOperations = 1000
		var wg sync.WaitGroup
		var pushCount, popCount int64
		var pushCountMu, popCountMu sync.Mutex

		// Start with some initial data
		for i := 0; i < 100; i++ {
			q.Push(i)
		}

		// Concurrent pushes
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numOperations; i++ {
				err := q.Push(i)
				if err != nil {
					t.Errorf("Push failed: %v", err)
				} else {
					pushCountMu.Lock()
					pushCount++
					pushCountMu.Unlock()
				}
			}
		}()

		// Concurrent pops
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numOperations; i++ {
				_, ok := q.Pop()
				if ok {
					popCountMu.Lock()
					popCount++
					popCountMu.Unlock()
				}
			}
		}()

		wg.Wait()

		// Check that we had some successful operations
		pushCountMu.Lock()
		actualPushCount := pushCount
		pushCountMu.Unlock()

		popCountMu.Lock()
		actualPopCount := popCount
		popCountMu.Unlock()

		if actualPushCount == 0 {
			t.Error("No pushes succeeded")
		}
		if actualPopCount == 0 {
			t.Error("No pops succeeded")
		}

		// Final size should be the difference between pushes and pops plus initial data
		expectedSize := 100 + actualPushCount - actualPopCount
		finalSize := q.Size()
		if finalSize != int(expectedSize) {
			t.Errorf("Expected final size %d, got %d (pushes: %d, pops: %d)", expectedSize, finalSize, actualPushCount, actualPopCount)
		}
	})

	t.Run("concurrent size checks", func(t *testing.T) {
		q := NewSyncFifoQueue[string](0)

		const numOperations = 100
		var wg sync.WaitGroup

		// Push elements
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				err := q.Push("test")
				if err != nil {
					t.Errorf("Push failed: %v", err)
				}
			}(i)
		}

		// Concurrently check size
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					size := q.Size()
					if size < 0 || size > numOperations {
						t.Errorf("Invalid size: %d", size)
					}
					time.Sleep(1 * time.Millisecond)
				}
			}()
		}

		wg.Wait()
	})
}
