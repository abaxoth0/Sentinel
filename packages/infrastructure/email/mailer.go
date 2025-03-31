package email

import (
	"context"
	"fmt"
	"log"
	"sentinel/packages/structs"
	"sync"
)

type Mailer struct {
    queue *structs.SyncFifoQueue[Mail]
    name string
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
            return nil
        default:
            mail, ok := m.queue.Pop()

            if !ok {
                m.queue.WaitTillNotEmpty()
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
    m.mut.Lock()

    if !m.isRunning {
        m.mut.Unlock()
        return fmt.Errorf("mailer '%s' is not running, hence can't be stopped", m.name)
    }

    m.isRunning = false

    m.mut.Unlock()

    m.queue.WaitTillEmpty()

    m.cancel()

    log.Printf(
        "[ EMAIL ] Mailer '%s' shut down with %d pending emails.\n",
        m.name, m.queue.Size(),
    )

    return nil
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

