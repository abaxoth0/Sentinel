package sessiontable

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
)

func NewRevokeAllUserSessionsQuery(act *ActionDTO.UserTargeted) *query.Query {
	return query.New(revokeAllUserSessionsSQL, act.TargetUID)
}

func (m *Manager) RevokeSession(act *ActionDTO.UserTargeted, sessionID string) *Error.Status {
	log.DB.Trace("Revoking user session...", nil)

	if act.RequesterUID != act.TargetUID {
		if err := authz.User.Logout(act.RequesterRoles); err != nil {
			return err
		}
	}

	session, err := m.getSessionByID(sessionID, false)
	if err != nil {
		return err
	}

	revokeQuery := query.New(
		`UPDATE "user_session" SET revoked_at = NOW() WHERE id = $1;`,
		sessionID,
	)

	audit := newAuditDTO(audit.DeleteOperation, &act.Basic, session)

	if err := execTxWithAudit(&audit, revokeQuery); err != nil {
		return err
	}

	cache.Client.Delete(
		cache.KeyBase[cache.SessionByID] + sessionID,
		cache.KeyBase[cache.UserBySessionID] + sessionID,
	)

	log.DB.Trace("Revoking user session: OK", nil)

	return nil
}

func (m *Manager) deleteSessionsCache(sessions []*SessionDTO.Full) *Error.Status {
	cacheKeys := make([]string, 0, len(sessions) * 2)

	for _, session := range sessions {
		cacheKeys = append(cacheKeys, cache.KeyBase[cache.SessionByID] + session.ID)
		cacheKeys = append(cacheKeys, cache.KeyBase[cache.UserBySessionID] + session.ID)
	}

	return cache.Client.Delete(cacheKeys...)
}

const revokeAllUserSessionsSQL = `UPDATE "user_session" SET revoked_at = NOW() WHERE user_id = $1;`

func (m *Manager) RevokeAllUserSessions(act *ActionDTO.UserTargeted) *Error.Status {
	log.DB.Trace("Revoking all user sessions...", nil)

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

	queries := make([]*query.Query, 0, len(sessions) + 1)
	queries = append(queries, revokeQuery)

	for i := range queries {
		auditDTO := newAuditDTO(audit.DeleteOperation, &act.Basic, sessions[i])
		queries = append(queries, newAuditQuery(&auditDTO))
	}

	if err := transaction.New(queries...).Exec(connection.Primary); err != nil {
		return err
	}

	m.deleteSessionsCache(sessions)

	log.DB.Trace("Revoking all user sessions: OK", nil)

	return nil
}

// Invalidates sessions cache of user with specified ID
func (m *Manager) DeleteUserSessionsCache(UID string) *Error.Status {
	log.DB.Trace("Deleting sessions cache for user "+UID+"...", nil)

	if e := validation.UUID(UID); e != nil {
		errMsg := "User ID must be a valid UUID: " + UID
		log.DB.Error("Failed to delete sessions cache for user "+UID, errMsg, nil)
		return Error.NewStatusError(errMsg, http.StatusBadRequest)
	}

	sessions, err := m.getUserSessions(UID)
	if err != nil {
		if err == Error.StatusNotFound {
			log.DB.Trace("Failed to delete sessions cache for user: "+UID+": User has no sessions", nil)
			return nil
		}
		log.DB.Error("Failed to delete sessions cache for user "+UID, err.Error(), nil)
		return err
	}

	if err := m.deleteSessionsCache(sessions); err != nil {
		log.DB.Error("Failed to delete sessions cache for user "+UID, err.Error(), nil)
		return err
	}

	log.DB.Trace("Deleting sessions cache for user: "+UID+":OK", nil)

	return nil
}

