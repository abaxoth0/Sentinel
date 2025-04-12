package email

import (
	"context"
	"fmt"
	"log"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"slices"
	"time"

	"gopkg.in/gomail.v2"
)

type Email interface {
    Send() *Error.Status
}

var MainMailer *Mailer
var dialer *gomail.Dialer
var isRunning = false

func Run() {
    if isRunning {
        panic("email module is already running")
    }

    initActivationEmailBody()

    validSMTPPorts := []int{587, 25, 465, 2525}

    if !slices.Contains(validSMTPPorts, config.Email.SmtpPort) {
        log.Fatalln("[ EMAIL ] Fatal error: invlid SMTP port")
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
        return fmt.Errorf("email module isn't running, hence can't be stopped")
    }

    defer func(){ isRunning = false }()

    if err := MainMailer.Stop(); err != nil {
        return err
    }

    return nil
}

