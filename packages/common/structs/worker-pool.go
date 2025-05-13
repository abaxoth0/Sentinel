package structs

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Task interface {
    Process()
}

type WorkerPool struct {
    canceled  atomic.Bool
    queue     *SyncFifoQueue[Task]
    waiter    Waiter
    ctx       context.Context
    cancel    context.CancelFunc
    wg        *sync.WaitGroup
    done      chan struct{}
}

// Creates new worker pool with specified waiter and parent context.
func NewWorkerPool(ctx context.Context, waiter Waiter) *WorkerPool {
    ctx, cancel := context.WithCancel(ctx)

    return &WorkerPool{
        queue: NewSyncFifoQueue[Task](),
        waiter: waiter,
        ctx: ctx,
        cancel: cancel,
        wg: new(sync.WaitGroup),
        done: make(chan struct{}),
    }
}

// Starts worker pool.
// Will process tasks in batches if 'batch' is true
func (wp *WorkerPool) Start(batch bool) error {
    if wp.canceled.Load() {
        return fmt.Errorf("worker pool is canceled")
    }

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
    var once sync.Once

    done := make(chan struct{})
    closeDone := func () {
        close(done)
    }

    go func() {
        for {
            select {
            case <-done:
                return
            default:
                if wp.queue.Size() >= batchSize {
                    once.Do(closeDone)
                    return
                }
                wp.waiter.Wait()
            }
        }
    }()

    // block till either there are will be enough
    // elements to procces batch, either timeout exceeded
    select {
    case<-done:
    case<-time.After(time.Millisecond * 50):
        once.Do(closeDone) // terminate waiting goroutine

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
    if wp.canceled.Load() {
        return fmt.Errorf("worker pool already canceled")
    }

    wp.cancel()
    wp.canceled.Store(true)
    wp.waiter.Wake()

    <-wp.done

    return nil
}

func (wp *WorkerPool) IsCanceled() bool {
    return wp.canceled.Load()
}

// Pushes a new task into a worker pool.
// Returns error on trying to push into a canceled worker pool
func (wp *WorkerPool) Push(t Task) error {
    if wp.canceled.Load() {
        return fmt.Errorf("can't push in canceled worker pool")
    }

    wp.queue.Push(t)
    wp.waiter.Wake() // notify waiters that there are a new task in queue

    return nil
}

