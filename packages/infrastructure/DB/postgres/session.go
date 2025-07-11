package postgres

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
)

// TODO add cache
type session struct {
	//
}

func (_ *session) SaveSession(session *SessionDTO.Full) *Error.Status {
	query := newQuery(
		`INSERT INTO "user_session" (id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, location, created_at, last_used_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`,
		session.ID,
		session.UserID,
		session.UserAgent,
		session.IpAddress,
		session.DeviceID,
		session.DeviceType,
		session.OS,
		session.OSVersion,
		session.Browser,
		session.BrowserVersion,
		session.Location,
		session.CreatedAt,
		session.LastUsedAt,
		session.ExpiresAt,
	)

	return query.Exec(primaryConnection)
}

func (_ *session) getSessionByID(sessionID string, revoked bool) (*SessionDTO.Full ,*Error.Status) {
	query := newQuery(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, location, created_at, last_used_at, expires_at, revoked FROM "user_session" WHERE id = $1 AND revoked = $2;`,
		sessionID, revoked,
	)

	dto, err := query.FullSessionDTO(replicaConnection)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (s *session) GetSessionByID(act *ActionDTO.Targeted, sessionID string, revoked bool) (*SessionDTO.Full ,*Error.Status) {
	if err := authz.User.GetUserSession(
		act.TargetUID == act.RequesterUID,
		act.RequesterRoles,
	); err != nil {
		return nil, err
	}

	return s.getSessionByID(sessionID, revoked)
}

func (_ *session) GetUserSessions(act *ActionDTO.Targeted) ([]*SessionDTO.Public, *Error.Status) {
	if err := authz.User.GetUserSession(
		act.TargetUID == act.RequesterUID,
		act.RequesterRoles,
	); err != nil {
		return nil, err
	}

	query := newQuery(
		`SELECT id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, location, created_at, last_used_at, expires_at FROM "user_session" WHERE user_id = $1 AND revoked = false;`,
		act.TargetUID,
	)

	sessions, err := query.CollectPublicSessionDTO(replicaConnection)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (_ *session) GetSessionByDeviceAndUserID(deviceID string, UID string) (*SessionDTO.Full ,*Error.Status) {
	query := newQuery(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, location, created_at, last_used_at, expires_at, revoked FROM "user_session" WHERE device_id = $1 AND user_id = $2 AND revoked = false;`,
		deviceID,
		UID,
	)

	dto, err := query.FullSessionDTO(replicaConnection)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (_ *session) UpdateSession(sessionID string, newSession *SessionDTO.Full) *Error.Status {
	query := newQuery(
		`UPDATE "user_session" SET
		user_id = $1, user_agent = $2, ip_address = $3, device_id = $4, device_type = $5, os = $6, os_version = $7, browser = $8, browser_version = $9, location = $10, last_used_at = $11, expires_at = $12
		WHERE id = $13 AND revoked = false;`,
		newSession.UserID,
		newSession.UserAgent,
		newSession.IpAddress,
		newSession.DeviceID,
		newSession.DeviceType,
		newSession.OS,
		newSession.OSVersion,
		newSession.Browser,
		newSession.BrowserVersion,
		newSession.Location,
		newSession.LastUsedAt,
		newSession.ExpiresAt,
		sessionID,
	)
	return query.Exec(primaryConnection)
}

func (s *session) RevokeSession(act *ActionDTO.Targeted, sessionID string) *Error.Status {
	if act.RequesterUID != act.TargetUID {
		if err := authz.User.Logout(act.RequesterRoles); err != nil {
			return err
		}
	}

	if _, err := s.getSessionByID(sessionID, false); err != nil {
		return err
	}

	query := newQuery(
		`UPDATE "user_session" SET revoked = true WHERE id = $1;`,
		sessionID,
	)

	cache.Client.Delete(
		cache.KeyBase[cache.SessionByID] + sessionID,
		cache.KeyBase[cache.UserBySessionID] + sessionID,
	)

	return query.Exec(primaryConnection)
}

