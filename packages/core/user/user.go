package user

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"sentinel/packages/core"
	"strings"
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
	passwordSize := len(strings.ReplaceAll(password, " ", ""))

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return invalidPasswordLength
    }

	return nil
}

func VerifyLogin(login string) *Error.Status {
    length := len(strings.ReplaceAll(login, " ", ""))

    if length < 5 || length > 72 {
        return Error.NewStatusError(
            "Логин должен находиться в диапозоне от 5 до 72 символов.",
            http.StatusBadRequest,
        )
    }

    if config.App.IsLoginEmail {
        if err := validation.Email(login); err != nil {
            // If err is not nil then it maybe only Error.InvalidValie,
            // cuz login was already checked for zero or whitespaces value
            return Error.NewStatusError(
                "Неверный логин: неподустимый формат E-Mail'а",
                http.StatusBadRequest,
            )
        }
    }

    return nil
}

