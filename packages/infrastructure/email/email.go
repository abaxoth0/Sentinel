package email

import (
	"context"
	"errors"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/validation"
	"slices"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

var log = logger.NewSource("EMAIL", logger.Default)

type EmailType int8

const (
	PasswordResetEmail EmailType = iota
	ActivationEmail
	PasswordChangeAlertEmail
	LoginChangeAlertEmail
	NewSessionAlertEmail
)

var emailsNames = map[EmailType]string{
	PasswordResetEmail:       "forgot pasword",
	ActivationEmail:          "activation",
	PasswordChangeAlertEmail: "password change alert",
	LoginChangeAlertEmail:    "login change alert",
	NewSessionAlertEmail:     "new session alert",
}

func (t EmailType) Name() (string, bool) {
	if name, ok := emailsNames[t]; ok {
		return name, true
	}
	return "", false
}

var emailsSubjects = map[EmailType]string{
	PasswordResetEmail:       "Password reset",
	ActivationEmail:          "Account activation",
	PasswordChangeAlertEmail: "Security Alert: password changed",
	LoginChangeAlertEmail:    "Security Alert: login changed",
	NewSessionAlertEmail:     "Security Alert: new sign-in",
}

func (t EmailType) Subject() (string, bool) {
	if sub, ok := emailsSubjects[t]; ok {
		return sub, true
	}
	return "", false
}

type AnyEmail interface {
	Type() EmailType
	Send() *Error.Status
	To() string
	Subject() string
}

func send(email AnyEmail, body string) *Error.Status {
	letter := gomail.NewMessage()

	letter.SetHeader("From", config.Secret.MailerEmail)
	letter.SetHeader("To", email.To())
	letter.SetHeader("Subject", email.Subject())
	letter.SetBody("text/html", body)

	if err := dialer.DialAndSend(letter); err != nil {
		return Error.NewStatusError(err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// Used for substituting placeholders with actual values in email body
type SubstitutionPlaceholder string

const (
	TokenPlaceholder    SubstitutionPlaceholder = "{{token}}"
	LocationPlaceholder SubstitutionPlaceholder = "{{location}}"
)

type Substitutions = map[SubstitutionPlaceholder]string

// Contains common fields for all emails: type, name, to and subject.
// Also implemets Type(), To() and Subject() methods of Email interface.
type Email struct {
	_type         EmailType
	name          string
	to            string
	subject       string
	substitutions Substitutions
}

func NewEmail(emailType EmailType, to string, substitutions Substitutions) (*Email, *Error.Status) {
	name, nameOk := emailType.Name()
	subject, subjectOk := emailType.Subject()
	if !nameOk || !subjectOk {
		log.Panic("Failed to create token email", "Invalid email name or subject", nil)
		return nil, Error.StatusInternalError
	}

	if err := validation.Email(to); err != nil {
		if err == Error.InvalidValue {
			errMsg := "Invlaid E-Mail format"
			log.Error("Failed to create "+name+" email", errMsg, nil)
			return nil, Error.NewStatusError(errMsg, http.StatusBadRequest)
		}
		if err == Error.NoValue {
			errMsg := "E-Mail is not specified"
			log.Error("Failed to create "+name+" email", errMsg, nil)
			return nil, Error.NewStatusError(errMsg, http.StatusBadRequest)
		}
	}

	return &Email{
		_type:         emailType,
		to:            to,
		name:          name,
		subject:       subject,
		substitutions: substitutions,
	}, nil
}

// Applies substitutions from subs to specified body.
// Returns empty string if fails.
func substitute(body string, placeholder SubstitutionPlaceholder, subs Substitutions) string {
	tk, ok := subs[placeholder]
	if !ok {
		log.Panic(
			"Failed to apply substitutions for email body",
			"Missing substitution for "+string(TokenPlaceholder)+" placeholder",
			nil,
		)
		return ""
	}
	var rawPlaceholder string
	if placeholder == TokenPlaceholder {
		rawPlaceholder = escapedTokenPlaceholder
	} else {
		rawPlaceholder = string(placeholder)
	}
	return strings.Replace(body, rawPlaceholder, tk, 1)
}

func (e *Email) Send() *Error.Status {
	var body string

	// Check if email type is valid and apply substitutions if needed
	switch e._type {
	case PasswordResetEmail:
		body = substitute(forgotPasswordEmailBody, TokenPlaceholder, e.substitutions)
	case ActivationEmail:
		body = substitute(activationEmailBody, TokenPlaceholder, e.substitutions)
	case PasswordChangeAlertEmail:
		body = passwordChangeAlertEmailBody
	case LoginChangeAlertEmail:
		body = loginChangeAlertEmailBody
	case NewSessionAlertEmail:
		body = substitute(newSessionAlertEmailBody, LocationPlaceholder, e.substitutions)
	default:
		log.Panic("Failed to send email", "Invalid email type", nil)
		return Error.StatusInternalError
	}
	// If body is empty that means substitute() failed
	if body == "" {
		return Error.StatusInternalError
	}

	log.Trace("Sending "+e.name+" email...", nil)

	if err := send(e, body); err != nil {
		log.Error("Failed to send "+e.name+" email", err.Error(), nil)
		return Error.NewStatusError(err.Error(), http.StatusInternalServerError)
	}

	log.Trace("Sending "+e.name+" email: OK", nil)

	return nil
}

func (e *Email) Type() EmailType {
	return e._type
}

func (e *Email) To() string {
	return e.to
}

func (e *Email) Subject() string {
	return e.subject
}

// Creates new template email, on success immediatly pushes it into the main mailer.
func EnqueueEmail(emailType EmailType, to string, subs Substitutions) *Error.Status {
	email, err := NewEmail(emailType, to, subs)
	if err != nil {
		return err
	}

	if err := MainMailer.Push(email); err != nil {
		return Error.StatusInternalError
	}

	return nil
}

var MainMailer *Mailer
var dialer *gomail.Dialer
var isInit = false

func Init() error {
	log.Info("Initializing email module...", nil)

	if isInit {
		errMsg := "Email module already initialized"
		log.Error("Failed to init email module", errMsg, nil)
		return errors.New(errMsg)
	}

	initTemplateEmailsBodies()

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

	isInit = true

	// give some time for MainMailer to start
	time.Sleep(time.Millisecond * 10)

	log.Info("Initializing email module: OK", nil)

	return nil
}

func Stop() error {
	log.Info("Stopping email module...", nil)

	if !isInit {
		errMsg := "email module isn't initialized, hence can't be stopped"
		log.Error("Failed to stop email module", errMsg, nil)
		return errors.New(errMsg)
	}

	if err := MainMailer.Stop(); err != nil {
		return err
	}

	log.Info("Stopping email module: OK", nil)

	return nil
}
