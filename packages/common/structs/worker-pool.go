package structs

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Task interface {
    Process()
}

type WorkerPool struct {
	canceled   atomic.Bool
    queue      *SyncFifoQueue[Task]
    ctx        context.Context
    cancel     context.CancelFunc
    wg         *sync.WaitGroup
	workerOnce sync.Once
	batchSize  int
}

// Creates new worker pool with specified waiter and parent context.
func NewWorkerPool(ctx context.Context, batchSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(ctx)

    return &WorkerPool{
        queue: NewSyncFifoQueue[Task](0),
        ctx: ctx,
        cancel: cancel,
        wg: new(sync.WaitGroup),
		batchSize: batchSize,
    }
}

// Starts worker pool.
// Will process tasks in batches if 'batch' is true
func (wp *WorkerPool) Start(workerCount int) {
	wp.workerOnce.Do(func() {
		for range workerCount {
			go wp.work()
		}
	})
}

func (wp *WorkerPool) work() {
	for {
		select {
		case <-wp.ctx.Done():
			for {
				tasks, ok := wp.queue.PopN(wp.batchSize)
				if !ok {
					return
				}
				wp.process(tasks)
			}
		default:
			if wp.queue.Size() == 0 {
				wp.queue.WaitTillNotEmpty(0)
				continue
			}

			tasks, ok := wp.queue.PopN(wp.batchSize)
			if !ok {
				continue
			}

			wp.process(tasks)
		}
	}
}

func (wp *WorkerPool) process(tasks []Task) {
	wp.wg.Add(1)

	if wp.batchSize == 1 {
		tasks[0].Process()
		wp.wg.Done()
		return
	}

	go func() {
		defer wp.wg.Done()
		for _, task := range tasks {
			task.Process()
		}
	}()
}

func (wp *WorkerPool) IsCanceled() bool {
	return wp.canceled.Load()
}

// Cancels worker pool.
// Worker pool will finish all its tasks before stopping.
// Once canceled, worker pool can't be started again.
func (wp *WorkerPool) Cancel() error {
	if wp.canceled.Load() {
		return errors.New("worker pool is already canceled")
	}

	wp.canceled.Store(true)
    wp.cancel()
	wp.wg.Wait()

	return nil
}

// Pushes a new task into a worker pool.
// Returns error on trying to push into a canceled worker pool
func (wp *WorkerPool) Push(t Task) error {
	if wp.canceled.Load() {
		return errors.New("can't push in canceled worker pool")
	}

	wp.queue.Push(t)

	return nil
}

