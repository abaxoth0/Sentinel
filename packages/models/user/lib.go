package user

import (
	"net/http"
	Error "sentinel/packages/errs"
)

func verifyPassword(password string) *Error.HTTP {
	passwordSize := len(password)

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return Error.NewHTTP("Недопустимый размер пароля. Пароль должен находится в диапозоне от 8 до 64 символов.", http.StatusBadRequest)
	}

	return nil
}
