package usertable

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
	"time"
)

func newAuditDTO(op audit.Operation, act *ActionDTO.UserTargeted, user *UserDTO.Basic) UserDTO.Audit {
    return UserDTO.Audit{
        ChangedUserID: act.TargetUID,
        ChangedByUserID: act.RequesterUID,
        Operation: string(op),
        ChangedAt: time.Now(),
		Reason: act.Reason,
		Basic: user,
    }
}

func newAuditQuery(dto *UserDTO.Audit) *query.Query {
    var deletedAt = util.Ternary(dto.IsDeleted(), &dto.DeletedAt, nil)
	var reason any = dto.Reason

	if dto.Reason == "" {
		reason = nil
	}

    return query.New(
        `INSERT INTO "audit_user"
        (changed_user_id, changed_by_user_id, operation, login, password, roles, deleted_at, changed_at, version, reason)
        VALUES
        ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
        dto.ChangedUserID,
        dto.ChangedByUserID,
        dto.Operation,
        dto.Login,
        dto.Password,
        dto.Roles,
        deletedAt,
        dto.ChangedAt,
		dto.Version,
		reason,
    )
}

func execTxWithAudit(dto *UserDTO.Audit, queries ...*query.Query) *Error.Status {
    queries = append(queries, newAuditQuery(dto))

    return transaction.New(queries...).Exec(connection.Primary)
}

