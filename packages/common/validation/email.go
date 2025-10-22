package validation

import (
	"regexp"
	Error "sentinel/packages/common/errors"
	"strings"
)

// Pretty close to RFC 5322 solution,
// but it's still not providing full features (like comments)
// and most likely will not handle all edge cases perfectly.
// But in this case, that's enough.
var emailPattern = regexp.MustCompile(`(?i)^(?:[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`)

func Email(email string) *Error.Validation {
	if strings.ReplaceAll(email, " ", "") == "" {
		return Error.NoValue
	}
	if !emailPattern.MatchString(email) {
		return Error.InvalidValue
	}
	return nil
}
