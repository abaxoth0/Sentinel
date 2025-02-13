package authentication

import (
	"log"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errs"
	"sentinel/packages/infrastructure/DB"

	"golang.org/x/crypto/bcrypt"
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

		return user, Error.NewStatusError("Неверный логин или пароль", http.StatusBadRequest)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// We know that on this stage only password is incorrect,
		// but there are no point to tell user about this (due to security reasons).
		return user, Error.NewStatusError("Неверный логин или пароль", http.StatusBadRequest)
	}

	return user, nil
}
