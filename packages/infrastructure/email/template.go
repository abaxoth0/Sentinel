package email

import (
	"bytes"
	"errors"
	"html/template"
)

// Parses given 'rawTemplate' replacing all placeholders
// via corresponding values of 'v'.
func parseEmailTemplate(rawTemplate string, v any) (string, error) {
	log.Info("Parsing email template...", nil)

	tmpl, err := template.New("email").Parse(rawTemplate)
	if err != nil {
		log.Error("Failed to parse email template", err.Error(), nil)
		return "", err
	}

	buf := new(bytes.Buffer)
	if e := tmpl.Execute(buf, v); e != nil {
		log.Error("Failed to parse email template", e.Error(), nil)
		return "", e
	}

	r := buf.String()
	if r == "<nil>" {
		errMsg := "Failed to read buffer"
		log.Error("Failed to parse email template", errMsg, nil)
		return "", errors.New(errMsg)
	}

	log.Info("Parsing email template: OK", nil)

	return string(r), nil
}
