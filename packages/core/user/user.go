package user

import (
	"net/http"
	Error "sentinel/packages/errors"
)

type Model struct {
	Login string
	Roles []string
	Password string
}

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
