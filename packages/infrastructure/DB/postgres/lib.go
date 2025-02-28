package postgres

import (
	"fmt"
	"sentinel/packages/core/user"
	Error "sentinel/packages/errors"

	"golang.org/x/crypto/bcrypt"
)

// TODO return string instead of byte array?
func hashPassword(password string) ([]byte, *Error.Status) {
	if err := user.VerifyPassword(password); err != nil {
		return nil, err
	}

	hashedPassword, e:= bcrypt.GenerateFromPassword([]byte(password), 12)

	if e != nil {
        fmt.Printf("[ ERROR ] Failed to generate hashed password: \n%v\n", e)

		return nil, Error.StatusInternalError
    }

    return hashedPassword, nil
}

