package authentication

import (
	"net/http"
	Error "sentinel/packages/common/errors"

	"golang.org/x/crypto/bcrypt"
)

var InvalidAuthCreditinals = Error.NewStatusError(
    "Неверный логин или пароль",
    http.StatusBadRequest,
)

// Comapres hashed password with it's possible plaintext equivalent.
// Returns nil on success, otherwise returns InvalidAuthCreditinals error.
func CompareHashAndPassword(hash string, password string) *Error.Status {
    e := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

    if e != nil {
        return InvalidAuthCreditinals
    }

    return nil
}

