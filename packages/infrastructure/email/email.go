package email

import (
	"context"
	"errors"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"slices"
	"time"

	"gopkg.in/gomail.v2"
)

var log = logger.NewSource("EMAIL", logger.Default)

type Email interface {
    Send() 		*Error.Status
	To()		string
	Subject() 	string
}

var MainMailer *Mailer
var dialer *gomail.Dialer
var isRunning = false

func Run() error {
	log.Info("Initializing email module...", nil)

    if isRunning {
		errMsg := "Mailer already running"
        log.Error("Failed to start mailer", errMsg, nil)
		return errors.New(errMsg)
    }

    initTokenEmailsBodies()

    validSMTPPorts := []int{587, 25, 465, 2525}

    if !slices.Contains(validSMTPPorts, config.Email.SmtpPort) {
		errMsg := "Invlid SMTP port"
        log.Error("Failed to start mailer", errMsg, nil)
		return errors.New(errMsg)
    }

    dialer = gomail.NewDialer(
        config.Email.SmtpHost,
        config.Email.SmtpPort,
        config.Secret.MailerEmail,
        config.Secret.MailerEmailPassword,
    )

	var err error

	MainMailer, err = NewMailer("main", context.Background(), nil)
	if err != nil {
		return err
	}

    go MainMailer.Run(5)

    isRunning = true

    // give some time for MainMailer to start
    time.Sleep(time.Millisecond * 10)

	log.Info("Initializing email module: OK", nil)

	return nil
}

func Stop() error {
	log.Info("Stopping email module...", nil)

    if !isRunning {
        return errors.New("email module isn't started, hence can't be stopped")
    }

    defer func(){ isRunning = false }()

    if err := MainMailer.Stop(); err != nil {
        return err
    }

	log.Info("Stopping email module: OK", nil)

    return nil
}

