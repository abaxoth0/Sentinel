package sessiontable

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
	"time"
)

func newAuditDTO(op audit.Operation, act *ActionDTO.Basic, session *SessionDTO.Full) SessionDTO.Audit {
    return SessionDTO.Audit{
        ChangedSessionID: session.ID,
        ChangedByUserID: act.RequesterUID,
        Operation: string(op),
        ChangedAt: time.Now(),
		Reason: act.Reason,
		Full: session,
    }
}

func newAuditQuery(dto *SessionDTO.Audit) *query.Query {
    // var revokedAt = util.Ternary(dto.IsRevoked(), &dto.RevokedAt, nil)
	var reason any = dto.Reason

	if dto.Reason == "" {
		reason = nil
	}

    return query.New(
        `INSERT INTO "audit_user_session"
        (changed_session_id, changed_by_user_id, operation, user_id, user_agent, ip_address, device_id, device_type, os, os_version, browser, browser_version, created_at, last_used_at, expires_at, revoked_at, changed_at, reason)
        VALUES
        ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
        dto.ChangedSessionID,
        dto.ChangedByUserID,
        dto.Operation,
		dto.UserID,
		dto.UserAgent,
		dto.IpAddress,
		dto.DeviceID,
		dto.DeviceType,
		dto.OS,
		dto.OSVersion,
		dto.Browser,
		dto.BrowserVersion,
		dto.CreatedAt,
		dto.LastUsedAt,
		dto.ExpiresAt,
		dto.RevokedAt,
        dto.ChangedAt,
		reason,
    )
}

func execTxWithAudit(dto *SessionDTO.Audit, queries ...*query.Query) *Error.Status {
    queries = append(queries, newAuditQuery(dto))

    return transaction.New(queries...).Exec(connection.Primary)
}

