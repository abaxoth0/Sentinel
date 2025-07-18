package sessiontable

import (
	Error "sentinel/packages/common/errors"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
)

func (m *Manager) UpdateSession(sessionID string, newSession *SessionDTO.Full) *Error.Status {
	updateQuery := query.New(
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

	return executor.Exec(connection.Primary, updateQuery)
}

