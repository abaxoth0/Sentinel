package authentication

import (
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/DB"

	"golang.org/x/crypto/bcrypt"
)

var invalidAuthCreditinals = Error.NewStatusError(
    "Неверный логин или пароль",
    http.StatusBadRequest,
)

// Returns indexedUser if auth data is correct, ExternalError otherwise.
func Login(login string, password string) (*UserDTO.Indexed, *Error.Status) {
	user, err := DB.Database.FindUserByLogin(login)

	if err != nil {
	    return user, err
    }

    e := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

    if e != nil {
		return user, invalidAuthCreditinals
	}

	return user, nil
}
