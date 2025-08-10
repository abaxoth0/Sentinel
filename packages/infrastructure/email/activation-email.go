package email

import (
	_ "embed"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api"
	"strings"

	rbac "github.com/abaxoth0/SentinelRBAC"
	"gopkg.in/gomail.v2"
)

//go:embed templates/activation-email.template.html
var activationEmailTemplate string

type activationEmailTemplateValues struct {
    ActivationURL string
}

var activationTokenPlaceholder = "{{token}}"

// Must be initialized via email.Run()
var activationEmailBody string

func initActivationEmailBody() {
    values := activationEmailTemplateValues{
        ActivationURL: api.GetBaseURL() + "/v1/user/activation/" + activationTokenPlaceholder,
    }

    b, err := parseTemplate(activationEmailTemplate, values)
    if err != nil {
        panic(err.Error())
    }

    activationEmailBody = b
}

type UserActivationEmail struct {
    To string
    Token string
}

// Creates new activation email, on success immediatly pushes it to mailer queue.
func CreateAndEnqueueActivationEmail(uid, email string) *Error.Status {
	log.Trace("Creating activation token...", nil)

	tk, err := token.NewActivationToken(
		uid,
		email,
		rbac.GetRolesNames(authz.Host.DefaultRoles),
	)
	if err != nil {
		log.Error("Failed to create new activation token", err.Error(), nil)
		return err
	}

	log.Trace("Creating activation token: OK", nil)

    activationEmail, err := NewUserActivationEmail(email, tk.String())
    if err != nil {
        return err
    }

	log.Trace("Pushing user activation email in mailer queue...", nil)

    if err := MainMailer.Push(activationEmail); err != nil {
        log.Error("Failed to push email in queue", err.Error(), nil)
        return Error.StatusInternalError
    }

	log.Trace("Pushing user activation email in mailer queue: OK", nil)

    return nil
}

func NewUserActivationEmail(to string, token string) (*UserActivationEmail, *Error.Status) {
    if err := validation.Email(to); err != nil {
        if err == Error.InvalidValue {
			errMsg := "Invlaid E-Mail format"
			log.Error("Failed to create user activation email", errMsg, nil)
            return nil, Error.NewStatusError(errMsg, http.StatusBadRequest)
        }
		if err == Error.NoValue {
			errMsg := "E-Mail is not specified"
			log.Error("Failed to create user activation email", errMsg, nil)
            return nil, Error.NewStatusError(errMsg, http.StatusBadRequest)
        }
    }

    return &UserActivationEmail{ To: to, Token: token }, nil
}

func (e *UserActivationEmail) Send() *Error.Status {
	log.Trace("Sending user activation email...", nil)

    email := gomail.NewMessage()

    email.SetHeader("From", config.Secret.MailerEmail)
    email.SetHeader("To", e.To)
    email.SetHeader("Subject", "Account activation")

    body := strings.Replace(activationEmailBody, activationTokenPlaceholder, e.Token, 1)

    email.SetBody("text/html", body)

    if err := dialer.DialAndSend(email); err != nil {
		log.Error("Failed to send user activation email", err.Error(), nil)
        return Error.NewStatusError(err.Error(), http.StatusInternalServerError)
    }

	log.Trace("Sending user activation email: OK", nil)

    return nil
}

