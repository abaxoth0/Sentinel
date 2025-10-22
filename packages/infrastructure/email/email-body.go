package email

import (
	_ "embed"
	"net/url"
	"sentinel/packages/common/config"
	"sentinel/packages/presentation/api"
)

var (
	//go:embed templates/login-change-alert-email.html
	loginChangeAlertEmailBody string
	//go:embed templates/password-change-alert-email.html
	passwordChangeAlertEmailBody string

	escapedTokenPlaceholder string = url.QueryEscape(string(TokenPlaceholder))

	//go:embed templates/password-reset-email.template.html
	passwordResetEmailTemplate string
	//go:embed templates/activation-email.template.html
	activationEmailTemplate string
	//go:embed templates/new-session-alert-email.template.html
	newSessionAlertEmailTemplate string

	// Must be initialized via email.Run()
	forgotPasswordEmailBody string
	// Must be initialized via email.Run()
	activationEmailBody string
	// Must be initialized via email.Run()
	newSessionAlertEmailBody string
)

func initTemplateEmailsBodies() {
	type activationEmailTemplateValues struct {
		ActivationURL string
	}

	activationEmailValues := activationEmailTemplateValues{
		ActivationURL: api.GetBaseURL() + "/v1/user/activation/" + string(TokenPlaceholder),
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
	query.Add("passwordResetToken", string(TokenPlaceholder))
	redirectURL.RawQuery = query.Encode()

	passwordResetEmailValues := passwordResetEmailTemplateValues{
		ResetPasswordURL: redirectURL.String(),
	}

	b, err = parseEmailTemplate(passwordResetEmailTemplate, passwordResetEmailValues)
	if err != nil {
		panic(err.Error())
	}

	forgotPasswordEmailBody = b

	type newSessionAlertEmailTemplateValues struct {
		Location string
	}

	newSessionAlertEmailValues := newSessionAlertEmailTemplateValues{
		Location: string(LocationPlaceholder),
	}

	b, err = parseEmailTemplate(newSessionAlertEmailTemplate, newSessionAlertEmailValues)
	if err != nil {
		panic(err.Error())
	}

	newSessionAlertEmailBody = b
}
