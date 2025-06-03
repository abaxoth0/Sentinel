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

var emailLogger = logger.NewSource("EMAIL", logger.Default)

type Email interface {
    Send() *Error.Status
}

var MainMailer *Mailer
var dialer *gomail.Dialer
var isRunning = false

func Run() {
    if isRunning {
        emailLogger.Fatal("Failed to start mailer", "Mailer already running", nil)
    }

    initActivationEmailBody()

    validSMTPPorts := []int{587, 25, 465, 2525}

    if !slices.Contains(validSMTPPorts, config.Email.SmtpPort) {
        emailLogger.Fatal("Failed to start mailer", "Invlid SMTP port", nil)
    }

    dialer = gomail.NewDialer(
        config.Email.SmtpHost,
        config.Email.SmtpPort,
        config.Secret.MailerEmail,
        config.Secret.MailerEmailPassword,
    )

    MainMailer = NewMailer("main", context.Background())

    go MainMailer.Run()

    isRunning = true

    // give some time for MainMailer to start
    time.Sleep(time.Millisecond * 50)
}

func Stop() error {
    if !isRunning {
        return errors.New("email module isn't running, hence can't be stopped")
    }

    defer func(){ isRunning = false }()

    if err := MainMailer.Stop(); err != nil {
        return err
    }

    return nil
}

