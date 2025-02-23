package authentication

import (
	"log"
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

	// TODO check this
	// If user was found (user.ID != "") and there are error, that means cursor closing failed. (see `findUserBy` method)
	// If user wasn't found and there are error, that means occured an unexpected error.
	if err != nil {
		if user.ID != "" {
			log.Printf("[ ERROR ] Failed to close cursor")
		}

	    return user, invalidAuthCreditinals
    }

    e := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

    if e != nil {
		return user, invalidAuthCreditinals
	}

	return user, nil
}
