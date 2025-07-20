package sessiontable

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
)

func NewRevokeAllUserSessionsQuery(act *ActionDTO.UserTargeted) *query.Query {
	return query.New(revokeAllUserSessionsSQL, act.TargetUID)
}

func (m *Manager) RevokeSession(act *ActionDTO.UserTargeted, sessionID string) *Error.Status {
	if act.RequesterUID != act.TargetUID {
		if err := authz.User.Logout(act.RequesterRoles); err != nil {
			return err
		}
	}

	if _, err := m.getSessionByID(sessionID, false); err != nil {
		return err
	}

	revokeQuery := query.New(
		`UPDATE "user_session" SET revoked = true WHERE id = $1;`,
		sessionID,
	)

	if err := executor.Exec(connection.Primary, revokeQuery); err != nil {
		return err
	}

	err := cache.Client.Delete(
		cache.KeyBase[cache.SessionByID] + sessionID,
		cache.KeyBase[cache.UserBySessionID] + sessionID,
	)
	if err != nil {
		sessionLogger.Error("Failed to delete cache", err.Error(), nil)
	}

	return nil
}

func (m *Manager) deleteSessionsCache(sessions []*SessionDTO.Public) *Error.Status {
	cacheKeys := make([]string, 0, len(sessions) * 2)

	for _, session := range sessions {
		cacheKeys = append(cacheKeys, cache.KeyBase[cache.SessionByID] + session.ID)
		cacheKeys = append(cacheKeys, cache.KeyBase[cache.UserBySessionID] + session.ID)
	}

	return cache.Client.Delete(cacheKeys...)
}

const revokeAllUserSessionsSQL = `UPDATE "user_session" SET revoked = true WHERE user_id = $1;`

func (m *Manager) RevokeAllUserSessions(act *ActionDTO.UserTargeted) *Error.Status {
	if act.RequesterUID != act.TargetUID {
		if err := authz.User.Logout(act.RequesterRoles); err != nil {
			return err
		}
	}

	sessions, err := m.getUserSessions(act.TargetUID)
	if err != nil {
		return err
	}

	revokeQuery := query.New(revokeAllUserSessionsSQL, act.TargetUID)

	if err := executor.Exec(connection.Primary, revokeQuery); err != nil {
		return err
	}

	if err := m.deleteSessionsCache(sessions); err != nil {
		sessionLogger.Error("Failed to delete sessions cache for user "+act.TargetUID, err.Error(), nil)
	}

	return nil
}

// Invalidates sessions cache of user with specified ID
func (m *Manager) DeleteUserSessionsCache(UID string) *Error.Status {
	sessionLogger.Trace("Deleting sessions cache for user "+UID+"...", nil)

	if e := validation.UUID(UID); e != nil {
		errMessage := "User ID must be a valid UUID: " + UID

		sessionLogger.Error("Failed to delete sessions cache for user "+UID, errMessage, nil)

		return Error.NewStatusError(errMessage, http.StatusBadRequest)
	}

	sessions, err := m.getUserSessions(UID)
	if err != nil {
		if err == Error.StatusNotFound {
			sessionLogger.Trace("Failed to delete sessions cache for user: "+UID+": User has no sessions", nil)
			return nil
		}
		sessionLogger.Error("Failed to delete sessions cache for user "+UID, err.Error(), nil)
		return err
	}

	if err := m.deleteSessionsCache(sessions); err != nil {
		sessionLogger.Error("Failed to delete sessions cache for user "+UID, err.Error(), nil)
		return err
	}

	sessionLogger.Trace("Deleting sessions cache for user: "+UID+":OK", nil)

	return nil
}

