package sessiontable

import (
	Error "sentinel/packages/common/errors"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"
)

func (m *Manager) SaveSession(session *SessionDTO.Full) *Error.Status {
	log.DB.Trace("Saving session...", nil)

	insertQuery := query.New(
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

	if err := executor.Exec(connection.Primary, insertQuery); err != nil {
		return err
	}

	log.DB.Trace("Saving session: OK", nil)

	return nil
}
