package postgres

import (
	Error "sentinel/packages/common/errors"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) ([]byte, *Error.Status) {
	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(password), 12)
	if e != nil {
        dbLogger.Error("Failed to generate hashed password", e.Error(), nil)

		return nil, Error.StatusInternalError
    }

    return hashedPassword, nil
}

