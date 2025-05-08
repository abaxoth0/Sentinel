package structs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Task interface {
    Process()
}

type WorkerPool struct {
    canceled  bool
    queue     *SyncFifoQueue[Task]
    waiter    Waiter
    ctx       context.Context
    cancel    context.CancelFunc
    m         *sync.Mutex
    wg        *sync.WaitGroup
    done      chan struct{}
}

// Creates new worker pool with specified waiter and parent context.
func NewWorkerPool(ctx context.Context, waiter Waiter) *WorkerPool {
    mut := new(sync.Mutex)
    ctx, cancel := context.WithCancel(ctx)

    return &WorkerPool{
        queue: NewSyncFifoQueue[Task](),
        waiter: waiter,
        ctx: ctx,
        cancel: cancel,
        m: mut,
        wg: new(sync.WaitGroup),
        done: make(chan struct{}),
    }
}

// Starts worker pool.
// Will process tasks in batches if 'batch' is true
func (wp *WorkerPool) Start(batch bool) error {
    wp.m.Lock()

    if wp.canceled {
        wp.m.Unlock()
        return fmt.Errorf("worker pool is canceled")
    }

    wp.m.Unlock()

    var process func()
    if batch {
        process = wp.processBatch
    } else {
        process = wp.processOne
    }

    for {
        select {
        case <-wp.ctx.Done():
            // Process all remain tasks before stopping.
            for wp.queue.Size() != 0 {
                process()
            }

            if batch {
                wp.wg.Wait()
            }

            close(wp.done)

            return nil
        default:
            process()
        }
    }
}

const batchSize int = 500

func (wp *WorkerPool) processBatch() {
    done := make(chan struct{})

    go func() {
        for {
            if wp.queue.Size() >= batchSize {
                close(done)
                return
            }
            wp.waiter.Wait()
        }
    }()

    // block till either there are will be enough
    // elements to procces batch, either timeout exceeded
    select {
    case<-done:
    case<-time.After(time.Millisecond * 50):
        if wp.queue.Size() == 0 {
            return
        }
    }

    batch := wp.queue.UnwrapAndFlush()

    // make sure that goroutine is started
    // (need to decrement counter before the end of goroutine work)
    wp.wg.Add(1)

    go func() {
        for _, task := range batch {
            wp.wg.Add(1)
            task.Process()
            wp.wg.Done()
        }

        wp.wg.Done()
    }()
}

// Proccesses one task from worker pool queue.
// If there are no task, then it will wait till task appears and return.
func (wp *WorkerPool) processOne() {
    task, ok := wp.queue.Pop()
    if !ok {
        wp.waiter.Wait()

        // Must return, cuz this function isn't tracking was context canceled or not.
        // (It's done in for-select inside of Start() method)
        // Even if it was canceled this function will never know about that and will wait forever.
        return
    }

    task.Process()
}

// Cancels worker pool.
// Worker pool will finish all its tasks before stopping.
// Once canceled, worker pool can't be started again.
func (wp *WorkerPool) Cancel() error {
    wp.m.Lock()
    defer wp.m.Unlock()

    if wp.canceled {
        return fmt.Errorf("worker pool already canceled")
    }

    wp.cancel()
    wp.canceled = true
    wp.waiter.Wake()

    <-wp.done

    return nil
}

func (wp *WorkerPool) Push(t Task) error {
    wp.m.Lock()

    if wp.canceled {
        wp.m.Unlock()
        return fmt.Errorf("can't push in canceled worker pool")
    }

    wp.m.Unlock()

    wp.queue.Push(t)
    wp.waiter.Wake() // notify waiters that there are a new task in queue

    return nil
}

