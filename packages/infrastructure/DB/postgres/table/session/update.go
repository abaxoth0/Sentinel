package sessiontable

import (
	Error "sentinel/packages/common/errors"
	actiondto "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/cache"
)

func (m *Manager) UpdateSession(act *actiondto.Basic, sessionID string, newSession *SessionDTO.Full) *Error.Status {
	log.DB.Trace("Updating session "+sessionID+"...", nil)

	session, err := m.getSessionByID(sessionID, false)
	if err != nil {
		return err
	}

	updateQuery := query.New(
		`UPDATE "user_session" SET
		user_id = $1, user_agent = $2, ip_address = $3, device_id = $4, device_type = $5, os = $6, os_version = $7, browser = $8, browser_version = $9, last_used_at = $10, expires_at = $11
		WHERE id = $12 AND revoked_at IS NULL;`,
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

	audit := newAuditDTO(audit.UpdatedOperation, act, session)

	if err := execTxWithAudit(&audit, updateQuery); err != nil {
		return err
	}

	cache.Client.Delete(
		cache.KeyBase[cache.SessionByID] + sessionID,
		cache.KeyBase[cache.UserBySessionID] + sessionID,
	)

	log.DB.Trace("Updating session "+sessionID+": OK", nil)

	return nil
}

