package authentication

import (
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/DB"

	"golang.org/x/crypto/bcrypt"
)

var invalidAuthCreditinals = Error.NewStatusError(
    "Неверный логин или пароль",
    http.StatusBadRequest,
)

// Comapres current password of user with ID == 'uid' with specified 'password'.
// If passwords hashes are equal - returns nil, otherwise returns *Error.Status.
func ComparePasswords(uid string, password string) *Error.Status {
    user, err := DB.Database.FindUserByID(uid)

    if err != nil {
        return err
    }

    e := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

    if e != nil {
        return Error.NewStatusError("Invalid password", http.StatusBadRequest)
    }

    return nil
}

func Login(login string, password string) (*UserDTO.Basic, *Error.Status) {
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

