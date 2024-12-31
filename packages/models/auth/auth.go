package auth

import (
	"net/http"
	ExternalError "sentinel/packages/errs"
	"sentinel/packages/models/search"

	"golang.org/x/crypto/bcrypt"
)

// Returns indexedUser if auth data is correct, ExternalError otherwise.
func Login(login string, password string) (*search.IndexedUser, *ExternalError.HTTP) {
	user, err := search.FindUserByLogin(login)

	// TODO check this
	// If user was found (user.ID != "") and there are error, that means cursor closing failed. (see `findUserBy` method)
	// If user wasn't found and there are error, that means occured an unexpected error.
	if err != nil && (user.ID != "") {
		return user, ExternalError.NewHTTP("Неверный логин или пароль", http.StatusBadRequest)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// We know that on this stage only password is incorrect,
		// but there are no point to tell user about this (due to security reasons).
		return user, ExternalError.NewHTTP("Неверный логин или пароль", http.StatusBadRequest)
	}

	return user, nil
}
