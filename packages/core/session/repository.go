package session

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
)

type Repository interface {
	creator
	seeker
	updater
	deleter
}

type creator interface {
	SaveSession(*SessionDTO.Full) *Error.Status
}

type seeker interface {
	GetSessionByID(act *ActionDTO.Targeted, sessionID string) (*SessionDTO.Full, *Error.Status)
	GetSessionByDeviceAndUserID(deviceID string, UID string) (*SessionDTO.Full, *Error.Status)
	GetUserSessions(act *ActionDTO.Targeted) ([]*SessionDTO.Public, *Error.Status)
}

type updater interface {
	UpdateSession(sessionID string, newSession *SessionDTO.Full) *Error.Status
}

type deleter interface {
	RevokeSession(act *ActionDTO.Targeted, sessionID string) *Error.Status
}

