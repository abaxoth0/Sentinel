package email

import (
	"context"
	"fmt"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/structs"
	"sync"
	"time"
)

type Mailer struct {
    queue *structs.SyncFifoQueue[Email]
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
        queue: structs.NewSyncFifoQueue[Email](0),
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

    emailLogger.Info("Mailer '"+m.name+"' started", nil)

    for {
        select {
        case <- m.ctx.Done():
            close(m.done) // notify waiters that mailer loop done it's work
            return nil
        default:
            email, ok := m.queue.PreserveAndPop()

            if !ok {
                // wait 1 second to avoid loading CPU too much
                // while there are no work.
                m.queue.WaitTillNotEmpty(time.Second)
                continue
            }

			emailLogger.Trace("Mailer '"+m.name+"': sending email...", nil)

            if err := email.Send(); err != nil {
                // Try to send email again if first attempt failed
                if err := email.Send(); err != nil {
					emailLogger.Error("Mailer '"+m.name+"': failed to send email", err.Error(), nil)
                }
            }

			emailLogger.Trace("Mailer '"+m.name+"': sending email: OK", nil)
        }
    }
}

// Stops mailer loop.
// Doesn't wait for all emails to be send.
func (m *Mailer) Stop() error {
	emailLogger.Info("Mailer '"+m.name+"': shutting down...", nil)

    m.mut.Lock()

    if !m.isRunning {
        m.mut.Unlock()
        return fmt.Errorf("mailer '%s' is not running, hence can't be stopped", m.name)
    }

    m.isRunning = false

    m.mut.Unlock()

	emailLogger.Info("Mailer '"+m.name+"': waiting till queue is empty...", nil)

    if timeout := m.queue.WaitTillEmpty(time.Second * 5); timeout != nil {
        emailLogger.Error(
			"Mailer '"+m.name+"': failed to wait till queue is empty",
			"Operation timeout waiting",
			nil,
		)
    } else {
		emailLogger.Info("Mailer '"+m.name+"': waiting till queue is empty: OK", nil)
    }

	emailLogger.Info("Mailer '"+m.name+"': waiting till current work is finished...", nil)

    // at this point mailer loop still can process some email so...
    m.cancel()

    for {
        select {
        // ...wait till mailer loop will finish it's current work...
        case <-m.done:
			emailLogger.Info("Mailer '"+m.name+"': waiting till current work is finished: OK", nil)
            emailLogger.Info(
				fmt.Sprintf("Mailer '"+m.name+"': shut down with %d pending emails", m.queue.Size()),
				nil,
            )

            return nil
        // ... or after some long time roll back queue and return timeout error.
        case <-time.After(time.Second * 5):
            emailLogger.Error(
				"Mailer '"+m.name+"': failed to wait till all current job is done, queue will be rolled back",
				"Operation timeout",
				nil,
			)

            m.queue.RollBack()

            return Error.StatusTimeout
        }
    }
}

// Returns all pending emails. Can't be called while mailer is running.
func (m *Mailer) Drain() ([]Email, error) {
    m.mut.Lock()
    defer m.mut.Unlock()

    if m.isRunning {
        return nil, fmt.Errorf("failed to drain emails from mailer '%s': mailer is running", m.name)
    }

    return m.queue.Unwrap(), nil
}

// Pushes new email to mailer queue
func (m *Mailer) Push(email Email) error {
    m.mut.Lock()
    defer m.mut.Unlock()

    if !m.isRunning {
        return fmt.Errorf("can't push to mailer '%s', mailer isn't running", m.name)
    }

    m.queue.Push(email)

    return nil
}

