package sessiontable

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
)

func (m *Manager) getSessionByID(sessionID string, revoked bool) (*SessionDTO.Full ,*Error.Status) {
	selectQuery := query.New(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked FROM "user_session" WHERE id = $1 AND revoked = $2;`,
		sessionID, revoked,
	)

	dto, err := executor.FullSessionDTO(connection.Replica, selectQuery)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (m *Manager) GetSessionByID(act *ActionDTO.UserTargeted, sessionID string, revoked bool) (*SessionDTO.Full ,*Error.Status) {
	if err := authz.User.GetUserSession(
		act.TargetUID == act.RequesterUID,
		act.RequesterRoles,
	); err != nil {
		return nil, err
	}

	return m.getSessionByID(sessionID, revoked)
}

func (m *Manager) getUserSessions(UID string) ([]*SessionDTO.Public, *Error.Status) {
	selectQuery := query.New(
		`SELECT id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at FROM "user_session" WHERE user_id = $1 AND revoked = false;`,
		UID,
	)

	sessions, err := executor.CollectPublicSessionDTO(connection.Replica, selectQuery)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (m *Manager) GetUserSessions(act *ActionDTO.UserTargeted) ([]*SessionDTO.Public, *Error.Status) {
	if err := authz.User.GetUserSession(
		act.TargetUID == act.RequesterUID,
		act.RequesterRoles,
		); err != nil {
		return nil, err
	}
	return m.getUserSessions(act.TargetUID)
}

func (m *Manager) GetSessionByDeviceAndUserID(deviceID string, UID string) (*SessionDTO.Full ,*Error.Status) {
	selectQuery := query.New(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked FROM "user_session" WHERE device_id = $1 AND user_id = $2 AND revoked = false;`,
		deviceID,
		UID,
	)

	dto, err := executor.FullSessionDTO(connection.Replica, selectQuery)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

