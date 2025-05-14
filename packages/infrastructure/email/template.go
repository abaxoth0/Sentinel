package email

import (
	"bytes"
	"errors"
	"html/template"
)

// Parses given 'rawTemplate' replacing all placeholders
// via corresponding values of 'v'.
func parseTemplate(rawTemplate string, v any) (string, error) {
    tmpl, err := template.New("email").Parse(rawTemplate)
    if err != nil {
        return "", err
    }

    buf := new(bytes.Buffer)
    if e := tmpl.Execute(buf, v); e != nil {
        return "", e
    }

    r := buf.String()
    if r == "<nil>" {
        return "", errors.New("failed to read buffer")
    }

    return string(r), nil
}

