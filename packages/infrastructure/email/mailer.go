package email

import (
	"context"
	"errors"
	"fmt"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/structs"
	"strconv"
	"sync/atomic"
)

type MailerOptions struct {
	// Default: 0
	MaxRetries int

	structs.WorkerPoolOptions
}

type Mailer struct {
	name      string
	isRunning atomic.Bool
	wp        *structs.WorkerPool
	ctx       context.Context
	cancel    context.CancelFunc
	opt       *MailerOptions
}

var mailesNames = map[string]bool{}

// Creates new mailer with specified name, parent context and options.
// If opt is nil then it will be created using default values of MailerOptions fields.
// Returns error if mailer with this name already exist.
func NewMailer(name string, ctx context.Context, opt *MailerOptions) (*Mailer, error) {
	if mailesNames[name] {
		return nil, errors.New("Mailer with name \"" + name + "\" already exist")
	}

	ctx, cancel := context.WithCancel(ctx)

	if opt == nil {
		opt = new(MailerOptions)
	}
	if opt.MaxRetries < 0 {
		opt.MaxRetries = 0
	}

	return &Mailer{
		name:   name,
		ctx:    ctx,
		cancel: cancel,
		wp:     structs.NewWorkerPool(ctx, &opt.WorkerPoolOptions),
		opt:    opt,
	}, nil
}

type emailTask struct {
	email  AnyEmail
	mailer *Mailer
}

func (t *emailTask) Process() {
	log.Trace("Mailer '"+t.mailer.name+"': sending email to "+t.email.To()+"...", nil)

	var err *Error.Status

	for i := range t.mailer.opt.MaxRetries + 1 {
		log.Trace("Attempts to send email to "+t.email.To()+": "+strconv.Itoa(i+1), nil)
		err = t.email.Send()
		if err == nil {
			break
		}
		log.Trace("Mailer '"+t.mailer.name+"': failed to send email to "+t.email.To()+", retrying", nil)
	}

	if err != nil {
		log.Error("Mailer '"+t.mailer.name+"': failed to send email to "+t.email.To(), err.Error(), nil)
		return
	}
	log.Trace("Mailer '"+t.mailer.name+"': email successfully sent to "+t.email.To(), nil)
}

// Starts mailer loop
func (m *Mailer) Run(workers int) error {
	if m.isRunning.Load() {
		return fmt.Errorf("mailer '%s' already running", m.name)
	}

	m.isRunning.Store(true)

	m.wp.Start(workers)

	log.Info("Mailer '"+m.name+"' started", nil)

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
func (m *Mailer) Push(email AnyEmail) error {
	if !m.isRunning.Load() {
		return fmt.Errorf("can't push to mailer '%s', mailer isn't running", m.name)
	}

	log.Trace("Pushing new email into the worker pool...", nil)

	err := m.wp.Push(&emailTask{
		email:  email,
		mailer: m,
	})
	if err != nil {
		log.Error("Failed to push new email into the worker pool", err.Error(), nil)
		return err
	}

	log.Trace("Pushing new email into the worker pool: OK", nil)

	return nil
}
