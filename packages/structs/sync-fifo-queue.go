package structs

import (
    "sync"
)

// TODO add size limit
// TODO add Peek() (non destructive way to read head element)

// Concurrency-safe first-in-first-out queue
type SyncFifoQueue[T any] struct {
    elems []T
    mut sync.Mutex
    cond *sync.Cond
}

func NewSyncFifoQueue[T any]() *SyncFifoQueue[T] {
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

// Deletes and returns first element of queue and true if queue isn't empty,
// if queue is empty returns zero-value of T and false.
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

/*
    IMPORTANT:
    DO NOT CALL THIS METHOD IN OTHER METHODS OF THIS STURCT,
    THIS WILL CAUSE DEADLOCK!!!
*/

// Returns amount of elements in queue
func (q *SyncFifoQueue[T]) Size() int {
    q.mut.Lock() // If mutex was locked before this line will cause deadlock, be careful
    l := len(q.elems)
    q.mut.Unlock()
    return l
}

// TODO Two functions below are pretty the main difference in conditions,
//      so maybe try to create a new function and use this function as wrappers for that?

// Waits till queue size is equal to 0.
func (q *SyncFifoQueue[T]) WaitTillEmpty() {
    q.mut.Lock()
    defer q.mut.Unlock()

    if len(q.elems) == 0 {
        return
    }

    for len(q.elems) != 0 {
        q.cond.Wait()
    }

}

// Waits till queue size is more then 0.
func (q *SyncFifoQueue[T]) WaitTillNotEmpty() {
    q.mut.Lock()

    if len(q.elems) > 0 {
        q.mut.Unlock()
        return
    }

    for len(q.elems) == 0 {
        q.cond.Wait()
    }
    q.mut.Unlock()
}

// Get copy of []T that is used by this queue under the hood
func (q *SyncFifoQueue[T]) Unwrap() []T {
    q.mut.Lock()

    r := make([]T, len(q.elems))

    copy(r, q.elems)

    q.mut.Unlock()

    return r
}

