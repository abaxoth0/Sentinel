package structs

import (
	"sync/atomic"
	"time"
)

const (
	BufferSize      = 1 << 16 // Power of 2
	BufferIndexMask = BufferSize - 1
)

var yieldWaiter = yieldSeqWait{}

type sequence struct {
    Value atomic.Int64
    // Padding to prevent false sharing
    _padding [56]byte
}

type seqWaiter interface {
    WaitFor(seq int64, cursor *sequence, done <-chan struct{})
}

type yieldSeqWait struct{}

func (w yieldSeqWait) WaitFor(seq int64, cursor *sequence, done <-chan struct{}) {
    for cursor.Value.Load() < seq {
        select {
        case <-done:
            return
        default:
            time.Sleep(10 * time.Microsecond)
        }
    }
}

type Disruptor[T any] struct {
	buffer  [BufferSize]T
	writer  sequence // write position (starts at -1)
	reader  sequence // read position (starts at -1)
	waiter  seqWaiter
	idle 	atomic.Bool
	done    chan struct{}
}

func NewDisruptor[T any]() *Disruptor[T] {
	d := &Disruptor[T]{
		done:   make(chan struct{}),
		waiter: yieldWaiter,
	}
	d.writer.Value.Store(-1)
	d.reader.Value.Store(-1)
	return d
}

func (d *Disruptor[T]) Close() {
	for !d.idle.Load() {
		time.Sleep(time.Microsecond * 10)
	}
	close(d.done)
}

func (d *Disruptor[T]) Publish(entry T) bool {
    select {
    case <-d.done:
        return false
    default:
        writer := d.writer.Value.Load()
        reader := d.reader.Value.Load()
        nextWriter := writer + 1

        // Check if buffer is full
        // NOTE: For buffer sizes ≤ 8, use (nextWriter - reader) > (BufferSize - 1)
        // to avoid off-by-one overwrites. For larger buffers (≥1024), the current check
        // is sufficient and more performant.
        if nextWriter - reader >= BufferSize {
            return false
        }

        d.buffer[nextWriter&BufferIndexMask] = entry
        d.writer.Value.Store(nextWriter)
        return true
    }
}

func (d *Disruptor[T]) Consume(handler func(T)) {
	var claimed int64 = d.reader.Value.Load() + 1
	closed := false

	for {
		select {
		case <-d.done:
			closed = true
		default:
			writer := d.writer.Value.Load()

			if claimed > writer {
				d.idle.Store(true)
				if closed {
					return
				}
				d.waiter.WaitFor(claimed, &d.writer, d.done)
				continue
			}

			d.idle.Store(false)

			// Process all entries from claimed to current writer
			for i := claimed; i <= writer; i++ {
				entry := d.buffer[i&BufferIndexMask]
				handler(entry)
				d.reader.Value.Store(i)
			}
			claimed = writer + 1 // Move to next unprocessed entry
		}
	}
}

