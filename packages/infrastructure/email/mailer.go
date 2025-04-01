package email

import (
	"context"
	"fmt"
	"log"
	Error "sentinel/packages/errors"
	"sentinel/packages/structs"
	"sync"
	"time"
)

type Mailer struct {
    queue *structs.SyncFifoQueue[Mail]
    name string
    done chan bool
    isRunning bool
    ctx context.Context
    cancel context.CancelFunc
    mut sync.Mutex
}

// Creates new mailer with specified name and parent context.
// (Several mailers can have the same names, be careful)
func NewMailer(name string, ctx context.Context) *Mailer {
    ctx, cancel := context.WithCancel(ctx)

    return &Mailer{
        name: name,
        ctx: ctx,
        cancel: cancel,
        queue: structs.NewSyncFifoQueue[Mail](),
        done: make(chan bool),
    }
}

// Starts mailer loop
func (m *Mailer) Run() error {
    m.mut.Lock()

    if m.isRunning {
        m.mut.Unlock()
        return fmt.Errorf("mailer '%s' already running", m.name)
    }

    m.isRunning = true
    m.mut.Unlock()

    log.Printf("[ EMAIL ] Mailer '%s' started.", m.name)

    for {
        select {
        case <- m.ctx.Done():
            close(m.done) // notify waiters that mailer loop done it's work
            return nil
        default:
            mail, ok := m.queue.PreserveAndPop()

            if !ok {
                // wait 1 second to avoid loading CPU too much
                // while there are no work.
                m.queue.WaitTillNotEmpty(time.Second)
                continue
            }

            if err := mail.Send(); err != nil {
                log.Printf("[ EMAIL ] Error sending email in mailer '%s': %v", m.name, err)
            }
        }
    }
}

// Stops mailer loop.
// Doesn't wait for all emails to be send.
func (m *Mailer) Stop() error {
    log.Printf("[ EMAIL ] Mailer '%s' is shutting down...\n", m.name)

    m.mut.Lock()

    if !m.isRunning {
        m.mut.Unlock()
        return fmt.Errorf("mailer '%s' is not running, hence can't be stopped", m.name)
    }

    m.isRunning = false

    m.mut.Unlock()

    log.Printf("[ EMAIL ] Mailer '%s' is waiting till mail queue is empty...\n", m.name)

    if timeout := m.queue.WaitTillEmpty(time.Second * 5); timeout != nil {
        log.Printf("[ EMAIL ] Mailer '%s': timeout waiting till queue is empty.\n", m.name)
    } else {
        log.Printf("[ EMAIL ] Mailer '%s' is waiting till mail queue is empty: OK\n", m.name)
    }

    log.Printf("[ EMAIL ] Mailer '%s' waiting till current work is finished...\n", m.name)

    // at this point mailer loop still can process some mail so...
    m.cancel()

    for {
        select {
        // ...wait till mailer loop will finish it's current work...
        case <-m.done:
            log.Printf("[ EMAIL ] Mailer '%s' waiting till current work is finished: OK\n", m.name)
            log.Printf(
                "[ EMAIL ] Mailer '%s' shut down with %d pending emails.\n",
                m.name, m.queue.Size(),
                )

            return nil
        // ... or after some long time.
        case <-time.After(time.Second * 5):
            log.Printf("[ EMAIL ] Mailer '%s': timeout waiting till current job is done. Rolling back queue.\n", m.name)

            m.queue.RollBack()

            return Error.StatusTimeout
        }
    }
}

// Returns all pending emails. Can't be called while mailer is running.
func (m *Mailer) Drain() ([]Mail, error) {
    m.mut.Lock()
    defer m.mut.Unlock()

    if m.isRunning {
        return nil, fmt.Errorf("failed to drain mails from mailer '%s': mailer is running", m.name)
    }

    return m.queue.Unwrap(), nil
}

// Pushes new mail to mailer queue
func (m *Mailer) Push(mail Mail) error {
    m.mut.Lock()
    defer m.mut.Unlock()

    if !m.isRunning {
        return fmt.Errorf("can't push to mailer '%s', mailer isn't running", m.name)
    }

    m.queue.Push(mail)

    return nil
}

