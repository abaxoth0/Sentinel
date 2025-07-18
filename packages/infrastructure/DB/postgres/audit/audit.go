package audit

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
	"time"
)

func NewUser(operation Operation, act *ActionDTO.UserTargeted, user *UserDTO.Basic) UserDTO.Audit {
    return UserDTO.Audit{
        ChangedUserID: act.TargetUID,
        ChangedByUserID: act.RequesterUID,
        Operation: string(operation),
        Login: user.Login,
        Password: user.Password,
        Roles: user.Roles,
        DeletedAt: user.DeletedAt,
        ChangedAt: time.Now(),
    }
}

func NewUserQuery(dto *UserDTO.Audit) *query.Query {
    var deletedAt = util.Ternary(dto.IsDeleted(), &dto.DeletedAt, nil)

    return query.New(
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

func ExecTxWithAuditUser(dto *UserDTO.Audit, queries ...*query.Query) *Error.Status {
    queries = append(queries, NewUserQuery(dto))

    return transaction.New(queries...).Exec(connection.Primary)
}

