package postgres

import (
	"database/sql"
	Error "sentinel/packages/common/errors"
	"time"

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

// Sets V equal to T if T is valid, otherwise sets V to time.Time{}
func setTime(V *time.Time, T sql.NullTime) {
    if T.Valid {
        *V = T.Time
    } else {
        *V = time.Time{}
    }
}

