package postgres

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"time"
)

/*
   Audit is using for storing users modifications histories.
   It works pretty simple, - before any user modification in DB
   row, that will be modified should be stored into audit_user with
   new property: changed_at
*/

var deleteOperation = "D"
var updatedOperation = "U"
var restoreOperation = "R"

func newAudit(operation string, act *ActionDTO.Targeted, user *UserDTO.Basic) UserDTO.Audit {
    return UserDTO.Audit{
        ChangedUserID: act.TargetUID,
        ChangedByUserID: act.RequesterUID,
        Operation: operation,
        Login: user.Login,
        Password: user.Password,
        Roles: user.Roles,
        DeletedAt: user.DeletedAt,
        ChangedAt: time.Now(),
    }
}

func newAuditQuery(dto *UserDTO.Audit) *query {
    var deletedAt = util.Ternary(dto.IsDeleted(), &dto.DeletedAt, nil)

    return newQuery(
        `INSERT INTO "audit_user"
        (changed_user_id, changed_by_user_id, operation, login, password, roles, deleted_at, changed_at)
        VALUES
        ($1, $2, $3, $4, $5, $6, $7, $8)`,
        dto.ChangedUserID,
        dto.ChangedByUserID,
        dto.Operation,
        dto.Login,
        dto.Password,
        dto.Roles,
        deletedAt,
        dto.ChangedAt,
    )
}

func execTransactionWithAudit(dto *UserDTO.Audit, queries ...*query) *Error.Status {
    queries = append(queries, newAuditQuery(dto))

    return newTransaction(queries...).Exec()
}

