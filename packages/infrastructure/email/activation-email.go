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

	rbac "github.com/StepanAnanin/SentinelRBAC"
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
        ActivationURL: api.GetBaseURL() + "/user/activation/" + activationTokenPlaceholder,
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
	emailLogger.Trace("Creating activation token...", nil)

	tk, err := token.NewActivationToken(
		uid,
		email,
		rbac.GetRolesNames(authz.Host.DefaultRoles),
	)
	if err != nil {
		emailLogger.Error("Failed to create new activation token", err.Error(), nil)
		return err
	}

	emailLogger.Trace("Creating activation token: OK", nil)

    activationEmail, err := NewUserActivationEmail(email, tk.String())
    if err != nil {
        return err
    }

    if err := MainMailer.Push(activationEmail); err != nil {
        emailLogger.Error("Failed to push email in queue", err.Error(), nil)
        return Error.StatusInternalError
    }

    return nil
}

func NewUserActivationEmail(to string, token string) (*UserActivationEmail, *Error.Status) {
    if err := validation.Email(to); err != nil {
        if err == Error.InvalidValue {
            return nil, Error.NewStatusError(
                "Invlaid E-Mail format",
                http.StatusBadRequest,
            )
        }
        if err == Error.NoValue {
            return nil, Error.NewStatusError(
                "E-Mail is not specified",
                http.StatusBadRequest,
            )
        }
    }

    return &UserActivationEmail{ To: to, Token: token }, nil
}

func (e *UserActivationEmail) Send() *Error.Status {
    email := gomail.NewMessage()

    email.SetHeader("From", config.Secret.MailerEmail)
    email.SetHeader("To", e.To)
    email.SetHeader("Subject", "Account activation")

    body := strings.Replace(activationEmailBody, activationTokenPlaceholder, e.Token, 1)

    email.SetBody("text/html", body)

    if err := dialer.DialAndSend(email); err != nil {
        return Error.NewStatusError(
            "Failed to send email: " + err.Error(),
            http.StatusInternalServerError,
        )
    }

    return nil
}

