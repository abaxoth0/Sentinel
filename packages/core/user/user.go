package user

import (
	"net/http"
	"regexp"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/core"
)

type Property core.EntityProperty

const (
    IdProperty Property = "id"
    LoginProperty Property = "login"
    RolesProperty Property = "roles"
    PasswordProperty Property = "password"
    DeletedAtProperty Property = "deletedAt"
    IsActiveProperty Property = "is_active"
)

// Represents user deletion state, might be:
// deleted (deletedState), not deleted (notDeletedState), any (anyState)
type State byte

const (
    NotDeletedState State = 0
    DeletedState State = 1
    AnyState State = 2
)

var invalidPasswordLength = Error.NewStatusError(
    "Пароль должен находится в диапозоне от 8 до 64 символов.",
    http.StatusBadRequest,
)

func VerifyPassword(password string) *Error.Status {
	passwordSize := len(password)

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return invalidPasswordLength
    }

	return nil
}

var invalidLoginLength = Error.NewStatusError(
    "Логин должен находиться в диапозоне от 5 до 72 символов.",
    http.StatusBadRequest,
)

var invalidEmailFormat = Error.NewStatusError(
    "Неверный логин: неподустимый формат E-Mail'а",
    http.StatusBadRequest,
)

// Pretty close to RFC 5322 solution,
// but it's still not providing full features (like comments)
// and most likely will not handle all edge cases perfectly.
// But in this case, that's enough.
var emailPattern = regexp.MustCompile(`(?i)^(?:[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`)

func VerifyLogin(login string) *Error.Status {
    if length := len(login); length < 5 || length > 72 {
        return invalidLoginLength
    }

    if config.App.IsLoginEmail {
        if !emailPattern.MatchString(login) {
            return invalidEmailFormat
        }
    }
    return nil
}

