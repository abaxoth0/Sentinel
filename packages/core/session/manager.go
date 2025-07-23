package session

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
)

type Manager interface {
	creator
	seeker
	updater
	deleter
}

type creator interface {
	SaveSession(*SessionDTO.Full) *Error.Status
}

type seeker interface {
	GetSessionByID(act *ActionDTO.UserTargeted, sessionID string) (*SessionDTO.Full, *Error.Status)
	GetRevokedSessionByID(act *ActionDTO.UserTargeted, sessionID string) (*SessionDTO.Full, *Error.Status)
	GetSessionByDeviceAndUserID(deviceID string, UID string) (*SessionDTO.Full, *Error.Status)
	GetUserSessions(act *ActionDTO.UserTargeted) ([]*SessionDTO.Public, *Error.Status)
}

type updater interface {
	UpdateSession(act *ActionDTO.Basic, sessionID string, newSession *SessionDTO.Full) *Error.Status
}

type deleter interface {
	RevokeSession(act *ActionDTO.UserTargeted, sessionID string) *Error.Status
	RevokeAllUserSessions(act *ActionDTO.UserTargeted) *Error.Status
	DeleteUserSessionsCache(UID string) *Error.Status
}

