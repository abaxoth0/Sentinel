package user

import (
	"net/http"
	ExternalError "sentinel/packages/error"
)

func verifyPassword(password string) *ExternalError.Error {
	passwordSize := len(password)

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return ExternalError.New("Недопустимый размер пароля. Пароль должен находится в диапозоне от 8 до 64 символов.", http.StatusBadRequest)
	}

	return nil
}
