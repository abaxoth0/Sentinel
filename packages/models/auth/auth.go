package auth

import (
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/search"

	"golang.org/x/crypto/bcrypt"
)

// Returns indexedUser if auth data is correct, ExternalError otherwise.
func Login(login string, password string) (*search.IndexedUser, *ExternalError.Error) {
	user, err := search.FindUserByLogin(login)

	// If user was found (user != indexedUser{}) and there are error, that means cursor closing failed. (see `findUserBy` method)
	// If user wasn't found and there are error, that means occured an unexpected error.
	if err != nil && (*user != search.IndexedUser{}) {
		return user, ExternalError.New("Неверный логин или пароль", http.StatusBadRequest)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// We know that on this stage incorrect only password,
		// but there are no point to tell user about this, due to security reasons.
		return user, ExternalError.New("Неверный логин или пароль", http.StatusBadRequest)
	}

	return user, nil
}
