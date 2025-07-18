package usertable

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	SessionTable "sentinel/packages/infrastructure/DB/postgres/table/session"

	"golang.org/x/crypto/bcrypt"
)

var userLogger = logger.NewSource("USER", logger.Default)

type Manager struct {
	session *SessionTable.Manager
}

func NewManager(session *SessionTable.Manager) *Manager {
	return &Manager{
		session: session,
	}
}

func hashPassword(password string) ([]byte, *Error.Status) {
	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(password), 12)
	if e != nil {
        userLogger.Error("Failed to generate password hash", e.Error(), nil)

		return nil, Error.StatusInternalError
    }

    return hashedPassword, nil
}

