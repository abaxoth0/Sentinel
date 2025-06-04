package structs

import (
    Error "sentinel/packages/common/errors"
    "sync"
    "time"
)

// TODO add size limit

// Concurrency-safe first-in-first-out queue
type SyncFifoQueue[T comparable] struct {
    elems []T
    mut sync.Mutex
    cond *sync.Cond
    preserved T
}

func NewSyncFifoQueue[T comparable]() *SyncFifoQueue[T] {
    q := new(SyncFifoQueue[T])

    q.cond = sync.NewCond(&q.mut)

    return q
}

// Appends v to the end of queue
func (q *SyncFifoQueue[T]) Push(v T) {
    q.mut.Lock()

    wasEmpty := len(q.elems) == 0

    q.elems = append(q.elems, v)

    q.mut.Unlock()

    if wasEmpty {
        q.cond.Broadcast()
    }
}

// If queue isn't empty - returns first element of queue and true.
// If queue is empty - returns zero-value of T and false.
func (q *SyncFifoQueue[T]) Peek() (T, bool) {
    q.mut.Lock()
    defer q.mut.Unlock()

    var v T

    if len(q.elems) == 0 {
        return v, false
    }

    return q.elems[0], true
}

// Same as Peek(), but also deletes first element in queue.
func (q *SyncFifoQueue[T]) Pop() (T, bool) {
    q.mut.Lock()
    defer q.mut.Unlock()

    var v T

    if len(q.elems) == 0 {
        return v, false
    }

    v = q.elems[0]
    q.elems = q.elems[1:]

    if len(q.elems) == 0 {
        q.cond.Broadcast()
    }

    return v, true
}

// Pops n elements from the queue.
// If n is greater then queue size, then to prevent panic n will be equated to the queue size.
// If queue is empty - returns nil and false.
func (q *SyncFifoQueue[T]) PopN(n int) ([]T, bool) {
    q.mut.Lock()
    defer q.mut.Unlock()

	s := make([]T, n)
	size := len(q.elems)

    if size == 0 {
        return nil, false
    }

	if n > size {
		n = size
	}

	s = q.elems[:n]
    q.elems = q.elems[n:]

    if size == 0 {
        q.cond.Broadcast()
    }

    return s, true
}

// Preserves the head element of the queue
func (q *SyncFifoQueue[T]) Preserve() {
    q.mut.Lock()
    defer q.mut.Unlock()

    if len(q.elems) == 0 {
        return
    }

    q.preserved = q.elems[0]
}

// Restores preserved element.
// Does nothing if no element was preserved.
func (q *SyncFifoQueue[T]) RollBack() {
    q.mut.Lock()
    defer q.mut.Unlock()

    var zero T

    if q.preserved == zero {
        return
    }

    swap := make([]T, len(q.elems) + 1)

    swap[0] = q.preserved
    q.preserved = zero

    for i, e := range q.elems {
        swap[i + 1] = e
    }

    q.elems = swap
}

// Do what is supposed by it's name:
// Just calls Preserve() and after that calls and returns Pop()
func (q *SyncFifoQueue[T]) PreserveAndPop() (T, bool) {
    q.Preserve()
    return q.Pop()
}

// Returns amount of elements in queue
func (q *SyncFifoQueue[T]) Size() int {
    q.mut.Lock() // If mutex was locked before this line will cause deadlock, be careful
    l := len(q.elems)
    q.mut.Unlock()
    return l
}

// If timeout <= 0: Waits till 'waitCond' returns true.
// If timeout > 0: Waits till either 'waitCond' returns true, either timeout exceeded.
func (q *SyncFifoQueue[T]) wait(timeout time.Duration, waitCond func() bool) *Error.Status {
    q.mut.Lock()
    defer q.mut.Unlock()

    if timeout <= 0 {
        for waitCond() {
            q.cond.Wait()
        }
        return nil
    }

    timer := time.NewTimer(timeout)
    defer timer.Stop()

    for waitCond() {
        done := make(chan bool)

        go func (){
            q.cond.Wait()
            close(done)
        }()

        select {
        case <-done:
        case <-timer.C:
            q.cond.Broadcast()
            /*
               IMPORTANT
                Need wait till q.cond.Wait() finish it's work,
                cuz it's unlocks mutex while waiting and lock it again before returning,
                so if q.cond.Wait() still waits that means mutext is unlocked.
                On this state may occur 2 type of erros:
                1) If mutex unlocking before returning from this function (which is currently so):
                   Attempt to unlock a mutex that is already unlocked by q.cond.Wait() will cause panic.
                2) If mutex isn't unlocking before returning:
                   q.cond.Wait() will lock it after finishing it's work and that will cause a deadlock.
            */
            <-done
            return Error.StatusTimeout
        }
    }

    return nil
}

// Waits till queue size is equal to 0.
// To disable timeout set it to <= 0.
// returns Error.TimeoutStatus if timeout exceeded, nil otherwise.
func (q *SyncFifoQueue[T]) WaitTillEmpty(timeout time.Duration) *Error.Status {
    q.mut.Lock()

    if len(q.elems) == 0 {
        q.mut.Unlock()
        return nil
    }

    q.mut.Unlock()

    return q.wait(timeout, func () bool { return len(q.elems) > 0 })
}

// Waits till queue size is more then 0.
// To disable timeout set it to <= 0.
// returns Error.TimeoutStatus if timeout exceeded, nil otherwise.
func (q *SyncFifoQueue[T]) WaitTillNotEmpty(timeout time.Duration) *Error.Status {
    q.mut.Lock()

    if len(q.elems) > 0 {
        q.mut.Unlock()
        return nil
    }

    q.mut.Unlock()

    return q.wait(timeout, func () bool { return len(q.elems) == 0 })
}

// Get copy of []T that is used by this queue under the hood
func (q *SyncFifoQueue[T]) Unwrap() []T {
    q.mut.Lock()

    r := make([]T, len(q.elems))

    copy(r, q.elems)

    q.mut.Unlock()

    return r
}

// Same as Unwrap, but also deletes all elements in queue
func (q *SyncFifoQueue[T]) UnwrapAndFlush() []T {
    q.mut.Lock()

    r := make([]T, len(q.elems))

    copy(r, q.elems)

    q.elems = []T{}

    q.mut.Unlock()

    return r
}

