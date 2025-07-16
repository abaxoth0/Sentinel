package postgres

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
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
		`INSERT INTO "user_session" (id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`,
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
		session.CreatedAt,
		session.LastUsedAt,
		session.ExpiresAt,
	)

	return query.Exec(primaryConnection)
}

func (_ *session) getSessionByID(sessionID string, revoked bool) (*SessionDTO.Full ,*Error.Status) {
	query := newQuery(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked FROM "user_session" WHERE id = $1 AND revoked = $2;`,
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

func (_ *session) getUserSessions(UID string) ([]*SessionDTO.Public, *Error.Status) {
	query := newQuery(
		`SELECT id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at FROM "user_session" WHERE user_id = $1 AND revoked = false;`,
		UID,
	)

	sessions, err := query.CollectPublicSessionDTO(replicaConnection)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *session) GetUserSessions(act *ActionDTO.Targeted) ([]*SessionDTO.Public, *Error.Status) {
	if err := authz.User.GetUserSession(
		act.TargetUID == act.RequesterUID,
		act.RequesterRoles,
		); err != nil {
		return nil, err
	}
	return s.getUserSessions(act.TargetUID)
}

func (_ *session) GetSessionByDeviceAndUserID(deviceID string, UID string) (*SessionDTO.Full ,*Error.Status) {
	query := newQuery(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked FROM "user_session" WHERE device_id = $1 AND user_id = $2 AND revoked = false;`,
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
		user_id = $1, user_agent = $2, ip_address = $3, device_id = $4, device_type = $5, os = $6, os_version = $7, browser = $8, browser_version = $9, last_used_at = $10, expires_at = $11
		WHERE id = $12 AND revoked = false;`,
		newSession.UserID,
		newSession.UserAgent,
		newSession.IpAddress,
		newSession.DeviceID,
		newSession.DeviceType,
		newSession.OS,
		newSession.OSVersion,
		newSession.Browser,
		newSession.BrowserVersion,
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

func (s *session) deleteSessionsCache(sessions []*SessionDTO.Public) *Error.Status {
	cacheKeys := make([]string, 0, len(sessions) * 2)

	for _, session := range sessions {
		cacheKeys = append(cacheKeys, cache.KeyBase[cache.SessionByID] + session.ID)
		cacheKeys = append(cacheKeys, cache.KeyBase[cache.UserBySessionID] + session.ID)
	}

	return cache.Client.Delete(cacheKeys...)
}

const revokeAllUserSessionsSQL = `UPDATE "user_session" SET revoked = true WHERE user_id = $1;`

func (s *session) RevokeAllUserSessions(act *ActionDTO.Targeted) *Error.Status {
	if act.RequesterUID != act.TargetUID {
		if err := authz.User.Logout(act.RequesterRoles); err != nil {
			return err
		}
	}

	sessions, err := s.getUserSessions(act.TargetUID)
	if err != nil {
		return err
	}

	query := newQuery(revokeAllUserSessionsSQL, act.TargetUID)

	if err := query.Exec(primaryConnection); err != nil {
		return err
	}

	if err := s.deleteSessionsCache(sessions); err != nil {
		dbLogger.Error("Failed to delete sessions cache for user "+act.TargetUID, err.Error(), nil)
	}

	return nil
}

// Invalidates sessions cache of user with specified ID
func (s *session) DeleteUserSessionsCache(UID string) *Error.Status {
	dbLogger.Trace("Deleting sessions cache for user "+UID+"...", nil)

	if e := validation.UUID(UID); e != nil {
		errMessage := "User ID must be a valid UUID: " + UID

		dbLogger.Error("Failed to delete sessions cache for user "+UID, errMessage, nil)

		return Error.NewStatusError(errMessage, http.StatusBadRequest)
	}

	sessions, err := s.getUserSessions(UID)
	if err != nil {
		if err == Error.StatusNotFound {
			dbLogger.Trace("Failed to delete sessions cache for user: "+UID+": User has no sessions", nil)
			return nil
		}
		dbLogger.Error("Failed to delete sessions cache for user "+UID, err.Error(), nil)
		return err
	}

	if err := s.deleteSessionsCache(sessions); err != nil {
		dbLogger.Error("Failed to delete sessions cache for user "+UID, err.Error(), nil)
		return err
	}

	dbLogger.Trace("Deleting sessions cache for user: "+UID+":OK", nil)

	return nil
}

