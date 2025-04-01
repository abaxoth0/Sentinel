package postgres

import (
	"database/sql"
	"fmt"
	"sentinel/packages/core/user"
	Error "sentinel/packages/common/errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) ([]byte, *Error.Status) {
	if err := user.VerifyPassword(password); err != nil {
		return nil, err
	}

	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(password), 12)

	if e != nil {
        fmt.Printf("[ ERROR ] Failed to generate hashed password: \n%v\n", e)

		return nil, Error.StatusInternalError
    }

    return hashedPassword, nil
}

// if T is valid set V to it, otherwise set V to time.Time{}
func setTime(V *time.Time, T sql.NullTime) {
    if T.Valid {
        *V = T.Time
    } else {
        *V = time.Time{}
    }
}

