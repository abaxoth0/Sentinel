package structs

import "sync"

// Wait strategy.
type Waiter interface {
    Wait()
    Wake()
}

// Uses sync.Cond for waiting.
type CondWaiter struct {
    mut  *sync.Mutex
    cond *sync.Cond
}

// Creates a new CondWaiter with specified mutex.
func NewCondWaiter(mut *sync.Mutex) *CondWaiter {
    return &CondWaiter{
        mut: mut,
        cond: sync.NewCond(mut),
    }
}

func (w *CondWaiter) Wait() {
    w.mut.Lock()
    defer w.mut.Unlock()

    w.cond.Wait()
}

func (w *CondWaiter) Wake() {
    w.cond.Broadcast()
}

