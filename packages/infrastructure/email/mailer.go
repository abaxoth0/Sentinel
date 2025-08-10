package email

import (
	"context"
	"errors"
	"fmt"
	"sentinel/packages/common/structs"
	"strconv"
	"sync/atomic"
)

type MailerOptions struct {
	// Default: 1
	BatchSize 	int
	// Default: 0
	MaxRetries 	int
}

type Mailer struct {
	name 		string
	isRunning 	atomic.Bool
	wp 			*structs.WorkerPool
    ctx 		context.Context
    cancel 		context.CancelFunc
	opt 		*MailerOptions
}

var mailesNames = map[string]bool{}

// Creates new mailer with specified name, parent context and options.
// If opt is nil then it will be created using default values of MailerOptions fields.
// Returns error if mailer with this name already exist.
func NewMailer(name string, ctx context.Context, opt *MailerOptions) (*Mailer, error) {
	if mailesNames[name] {
		return nil, errors.New("Mailer with name \""+name+"\" already exist")
	}

    ctx, cancel := context.WithCancel(ctx)

	if opt == nil {
		opt = &MailerOptions{
			MaxRetries: 0,
			BatchSize: 1,
		}
	}
	if opt.BatchSize <= 0 {
		opt.BatchSize = 1
	}
	if opt.MaxRetries < 0 {
		opt.MaxRetries = 0
	}

    return &Mailer{
        name: name,
        ctx: ctx,
        cancel: cancel,
		wp: structs.NewWorkerPool(ctx, opt.BatchSize),
		opt: opt,
    }, nil
}

type emailTask struct {
	email 	Email
	mailer	*Mailer
}

func (t *emailTask) Process() {
	log.Trace("Mailer '"+t.mailer.name+"': sending email...", nil)

	for i := range t.mailer.opt.MaxRetries + 1 {
		log.Trace("Attempts to send email: "+strconv.Itoa(i+1), nil)
		err := t.email.Send()
		if err == nil {
			break
		}
		log.Error("Mailer '"+t.mailer.name+"': failed to send email", err.Error(), nil)
	}

	log.Trace("Mailer '"+t.mailer.name+"': sending email: OK", nil)
}

// Starts mailer loop
func (m *Mailer) Run(workers int) error {
    if m.isRunning.Load() {
        return fmt.Errorf("mailer '%s' already running", m.name)
    }

    m.isRunning.Store(true)

    log.Info("Mailer '"+m.name+"' started", nil)

	m.wp.Start(workers)

	return nil
}

// Stops mailer loop.
// Doesn't wait for all emails to be send.
func (m *Mailer) Stop() error {
	log.Info("Mailer '"+m.name+"': shutting down...", nil)

    if !m.isRunning.Load() {
        return fmt.Errorf("mailer '%s' is not running, hence can't be stopped", m.name)
    }

    m.isRunning.Store(false)

	if err := m.wp.Cancel(); err != nil {
		log.Error("Mailer '"+m.name+"': failed to shut down", err.Error(), nil)
		return err
	}

	log.Info("Mailer '"+m.name+"': shutting down: OK", nil)

	return nil
}

// Pushes new email to mailer queue
func (m *Mailer) Push(email Email) error {
    if !m.isRunning.Load() {
        return fmt.Errorf("can't push to mailer '%s', mailer isn't running", m.name)
    }

	log.Trace("Pushing new email into the worker pool...", nil)

	err := m.wp.Push(&emailTask{
		email: email,
		mailer: m,
	})
	if err != nil {
		log.Error("Failed to push new email into the worker pool", err.Error(), nil)
		return err
	}

	log.Trace("Pushing new email into the worker pool: OK", nil)

    return nil
}

