package user

import (
	"net/http"
	Error "sentinel/packages/errs"
)

type Raw struct {
	Login string
	Roles []string
}

type RawSecured struct {
	Password string
	Raw
}

func VerifyPassword(password string) *Error.Status {
	passwordSize := len(password)

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return Error.NewStatusError("Недопустимый размер пароля. Пароль должен находится в диапозоне от 8 до 64 символов.", http.StatusBadRequest)
	}

	return nil
}
