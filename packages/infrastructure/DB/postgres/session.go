package postgres

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
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

func (_ *session) GetSessionByID(sessionID string) (*SessionDTO.Full ,*Error.Status) {
	query := newQuery(
		`SELECT id, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, location, created_at, last_used_at, expires_at, revoked FROM "user_session" WHERE id = $1 AND revoked = false;`,
		sessionID,
	)

	dto, err := query.FullSessionDTO(replicaConnection)
	if err != nil {
		return nil, err
	}

	return dto, nil
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

func (_ *session) RevokeSession(act *ActionDTO.Targeted, sessionID string) *Error.Status {
	query := newQuery(
		`UPDATE "user_session" SET revoked = true WHERE id = $1;`,
		sessionID,
	)
	return query.Exec(primaryConnection)
}

