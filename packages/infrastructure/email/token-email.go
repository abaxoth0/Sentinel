package email

import (
	_ "embed"
	"net/http"
	"net/url"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api"
	"strings"

	"gopkg.in/gomail.v2"
)

const tokenPlaceholder string = "{{token}}"

var (
	escapedTokenPlaceholder string = url.QueryEscape(tokenPlaceholder)

	//go:embed templates/password-reset-email.template.html
	passwordResetEmailTemplate string
	//go:embed templates/activation-email.template.html
	activationEmailTemplate string

	// Must be initialized via email.Run()
	forgotPasswordEmailBody string
	// Must be initialized via email.Run()
	activationEmailBody string
)

func initTokenEmailsBodies() {
	type activationEmailTemplateValues struct {
		ActivationURL string
	}

    activationEmailValues := activationEmailTemplateValues{
        ActivationURL: api.GetBaseURL() + "/v1/user/activation/" + tokenPlaceholder,
    }

    b, err := parseEmailTemplate(activationEmailTemplate, activationEmailValues)
    if err != nil {
        panic(err.Error())
    }

    activationEmailBody = b

	type passwordResetEmailTemplateValues struct {
		ResetPasswordURL string
	}

	redirectURL, err := url.Parse(config.App.PasswordResetRedirectURL)
	if err != nil {
		panic(err.Error())
	}

	query := redirectURL.Query()
	query.Add("passwordResetToken", tokenPlaceholder)
	redirectURL.RawQuery = query.Encode()

    passwordResetEmailValues := passwordResetEmailTemplateValues{
        ResetPasswordURL: redirectURL.String(),
    }

    b, err = parseEmailTemplate(passwordResetEmailTemplate, passwordResetEmailValues)
    if err != nil {
        panic(err.Error())
    }

    forgotPasswordEmailBody = b
}

type TokenEmailType uint8

const (
	PasswordResetTokenType TokenEmailType = iota
	ActivationTokenType
)

func (t TokenEmailType) Name() (string, bool) {
	switch t {
	case PasswordResetTokenType:
		return "forgot-pasword", true
	case ActivationTokenType:
		return "user activation", true
	default:
		return "", false
	}
}

func (t TokenEmailType) Subject() (string, bool) {
	switch t {
	case PasswordResetTokenType:
		return "Password reset ", true
	case ActivationTokenType:
		return "Account activation", true
	default:
		return "", false
	}
}

type TokenEmail struct {
	tokenType 	TokenEmailType
    Token 		string

	plainEmail
}

func NewTokenEmail(tokenType TokenEmailType, to string, token string) (*TokenEmail, *Error.Status) {
	name, nameOk := tokenType.Name()
	subject, subjectOk := tokenType.Subject()
	if !nameOk || !subjectOk {
		log.Panic("Failed to create TokenEmail", "Unknown token type", nil)
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

	return &TokenEmail{
		Token: token,
		tokenType: tokenType,
		plainEmail: plainEmail{
			to: to,
			name: name,
			subject: subject,
		},
	}, nil
}

func (e *TokenEmail) Send() *Error.Status {
	var rawBody 	string

	switch e.tokenType {
	case PasswordResetTokenType:
			rawBody = forgotPasswordEmailBody
	case ActivationTokenType:
			rawBody = activationEmailBody
	default:
		log.Panic("Failed to send TokenEmail", "Unknown token type", nil)
		return Error.StatusInternalError
	}

	log.Trace("Sending "+e.name+" email...", nil)

    email := gomail.NewMessage()

    email.SetHeader("From", config.Secret.MailerEmail)
    email.SetHeader("To", e.to)
    email.SetHeader("Subject", e.subject)

    body := strings.Replace(rawBody, escapedTokenPlaceholder, e.Token, 1)

    email.SetBody("text/html", body)

    if err := dialer.DialAndSend(email); err != nil {
		log.Error("Failed to send "+e.name+" email", err.Error(), nil)
        return Error.NewStatusError(err.Error(), http.StatusInternalServerError)
    }

	log.Trace("Sending "+e.name+" email: OK", nil)

    return nil
}

// Creates new email, on success immediatly pushes it into the main mailer.
func EnqueueTokenEmail(tokenType TokenEmailType, uid string, email string) *Error.Status {
	var tk *token.SignedToken
	var err *Error.Status

	switch tokenType{
	case PasswordResetTokenType:
		tk, err = token.NewPasswordResetToken(uid, email)
	case ActivationTokenType:
		tk, err = token.NewActivationToken(uid, email)
	default:
		log.Panic("Failed to enqueue TokenEmail", "Unknown token type", nil)
		return Error.StatusInternalError
	}
	if err != nil {
		return err
	}

    letter, err := NewTokenEmail(tokenType, email, tk.String())
    if err != nil {
        return err
    }

    if err := MainMailer.Push(letter); err != nil {
        return Error.StatusInternalError
    }

    return nil
}

