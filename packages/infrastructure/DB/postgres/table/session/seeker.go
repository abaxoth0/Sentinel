package sessiontable

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
)

func (m *Manager) getSessionByID(sessionID string, revoked bool) (*SessionDTO.Full ,*Error.Status) {
	cond := util.Ternary(revoked, "IS NOT", "IS")

	selectQuery := query.New(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked_at FROM "user_session" WHERE id = $1 AND revoked_at `+cond+" NULL;",
		sessionID,
	)

	var cacheKey string
	if revoked {
		cacheKey = cache.KeyBase[cache.RevokedSessionByID] + sessionID
	} else {
		cacheKey = cache.KeyBase[cache.SessionByID] + sessionID
	}

	dto, err := executor.FullSessionDTO(connection.Replica, selectQuery, cacheKey)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (m *Manager) GetSessionByID(act *ActionDTO.UserTargeted, sessionID string) (*SessionDTO.Full ,*Error.Status) {
	if err := authz.User.GetUserSession(
		act.TargetUID == act.RequesterUID,
		act.RequesterRoles,
	); err != nil {
		return nil, err
	}

	return m.getSessionByID(sessionID, false)
}

func (m *Manager) GetRevokedSessionByID(act *ActionDTO.UserTargeted, sessionID string) (*SessionDTO.Full ,*Error.Status) {
	if err := authz.User.GetUserSession(
		act.TargetUID == act.RequesterUID,
		act.RequesterRoles,
	); err != nil {
		return nil, err
	}

	return m.getSessionByID(sessionID, true)
}

func (m *Manager) getUserSessions(UID string) ([]*SessionDTO.Full, *Error.Status) {
	selectQuery := query.New(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked_at FROM "user_session" WHERE user_id = $1 AND revoked_at IS NULL;`,
		UID,
	)

	sessions, err := executor.CollectFullSessionDTO(connection.Replica, selectQuery)
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

	fullSessions, err := m.getUserSessions(act.TargetUID)

	if err != nil {
		return nil, err
	}

	sessions := make([]*SessionDTO.Public, len(fullSessions))

	for i, fulllSession := range fullSessions {
		sessions[i] = fulllSession.MakePublic()
	}

	return sessions, nil
}

func (m *Manager) GetSessionByDeviceAndUserID(deviceID string, UID string) (*SessionDTO.Full ,*Error.Status) {
	selectQuery := query.New(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked_at FROM "user_session" WHERE device_id = $1 AND user_id = $2 AND revoked_at IS NULL;`,
		deviceID,
		UID,
	)

	dto, err := executor.FullSessionDTO(
		connection.Replica,
		selectQuery,
		cache.KeyBase[cache.SessionByDeviceAndUserID] + deviceID + "|" + UID,
	)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

