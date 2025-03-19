package user

import (
	"net/http"
	Error "sentinel/packages/errors"
)

// Represents one of user's properties, excluding password.
//
// To avoid possible vulnerabilities like SQL-injections
// all data of this type must be a predefined consts.
// Doing so there are no need in property validation cuz
// all properties are predefined and correct.
type Property string

const (
    IdProperty Property = "id"
    LoginProperty Property = "login"
    RolesProperty Property = "roles"
    PasswordProperty Property = "password"
    DeletedAtProperty Property = "deletedAt"
)

// Represents user deletion state, might be:
// deleted (deletedState), not deleted (notDeletedState), any (anyState)
type State byte

const (
    NotDeletedState State = 0
    DeletedState State = 1
    AnyState State = 2
)

func VerifyPassword(password string) *Error.Status {
	passwordSize := len(password)

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return Error.NewStatusError(
            "Пароль должен находится в диапозоне от 8 до 64 символов.",
            http.StatusBadRequest,
        )
	}

	return nil
}

