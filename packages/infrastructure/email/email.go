package email

import (
	"context"
	"log"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"slices"

	"gopkg.in/gomail.v2"
)

type Email interface {
    Send() *Error.Status
}

var MainMailer *Mailer
var dialer *gomail.Dialer

func Init() {
    createActivationEmailBody()

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
}

