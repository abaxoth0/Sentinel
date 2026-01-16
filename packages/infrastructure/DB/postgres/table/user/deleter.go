package usertable

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/dblog"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	SessionTable "sentinel/packages/infrastructure/DB/postgres/table/session"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	"slices"
	"strings"
	"time"
)

func (m *Manager) SoftDelete(act *ActionDTO.UserTargeted) *Error.Status {
	dblog.Logger.Info("Soft deleting user...", nil)

	if err := act.ValidateUIDs(); err != nil {
		dblog.Logger.Error("Failed to soft delete user", err.Error(), nil)
		return err
	}

	if err := authz.User.SoftDeleteUser(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

	user, err := m.GetUserByID(act.TargetUID)
	if err != nil {
		return err
	}

	auditUserDTO := newAuditDTO(audit.DeleteOperation, act, user)

	updateQuery := query.New(
		`UPDATE "user" SET deleted_at = $1, version = version + 1
        WHERE id = $2 AND deleted_at IS NULL;`,
		// deleted_at set manualy instead of using NOW()
		// cuz changed_at and deleted_at should be synchronized
		auditUserDTO.ChangedAt, act.TargetUID,
	)
	sessionsDeleteQuery := SessionTable.NewRevokeAllUserSessionsQuery(act)

	if err := execTxWithAudit(&auditUserDTO, updateQuery, sessionsDeleteQuery); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = &auditUserDTO.ChangedAt
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	m.session.DeleteUserSessionsCache(user.ID)

	dblog.Logger.Info("Soft deleting user: OK", nil)

	return nil
}

func (m *Manager) Restore(act *ActionDTO.UserTargeted, newLogin string) *Error.Status {
	dblog.Logger.Info("Restoring soft deleted user...", nil)

	if err := act.ValidateUIDs(); err != nil {
		dblog.Logger.Error("Failed to restore soft deleted user", err.Error(), nil)
		return err
	}

	if err := authz.User.RestoreUser(act.RequesterRoles); err != nil {
		return err
	}

	user, err := m.GetSoftDeletedUserByID(act.TargetUID)
	if err != nil {
		return err
	}

	auditUserDTO := newAuditDTO(audit.RestoreOperation, act, user)

	queries := []*query.Query{}

	err = m.checkLoginAvailability(user.Login)
	if err != nil {
		if err != loginAlreadyInUse {
			return err
		}

		if newLogin == "" {
			return Error.NewStatusError(
				"Login of this account has been taken by other user. To restore this account you need to specify a new login for it",
				http.StatusConflict,
			)
		}

		queries = append(queries, query.New(
			`UPDATE "user" SET login = $1 WHERE id = $2;`,
			newLogin, act.TargetUID,
		))
	}

	queries = append(queries,
		query.New(
			`UPDATE "user" SET deleted_at = NULL, version = version + 1
			WHERE id = $1 AND deleted_at IS NOT NULL;`,
			act.TargetUID,
		),
	)

	if err := execTxWithAudit(&auditUserDTO, queries...); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = new(time.Time)
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	dblog.Logger.Info("Restoring soft deleted user: OK", nil)

	return nil
}

func (m *Manager) bulkStateUpdate(newState user.State, act *ActionDTO.Basic, UIDs []string) *Error.Status {
	if newState != user.DeletedState && newState != user.NotDeletedState {
		dblog.Logger.Panic(
			"Invalid bulkStateUpdate call",
			"newState must be either user.DeletedState, either user.NotDeletedState",
			nil,
		)
		return Error.StatusInternalError
	}

	if err := act.ValidateRequesterUID(); err != nil {
		return err
	}

	for _, uid := range UIDs {
		if uid == act.RequesterUID {
			return Error.NewStatusError(
				util.Ternary(
					newState == user.DeletedState,
					"Can't self-delete in bulk operation",
					"Can't self-restore in bulk operation",
				),
				http.StatusBadRequest,
			)
		}
		if err := validation.UUID(uid); err != nil {
			return err.ToStatus(
				"One of user IDs has no value", // empty string or just a bunch of ' '
				"Invalid user ID format (expected UUID): "+uid,
			)
		}
	}

	cond := util.Ternary(newState == user.DeletedState, "IS", "IS NOT")
	selectQuery := query.New(
		`SELECT id, login, password, roles, deleted_at, created_at, version FROM "user" WHERE id = ANY($1) and deleted_at `+cond+` NULL;`,
		UIDs,
	)

	deletedUsers, err := executor.CollectFullUserDTO(connection.Replica, selectQuery)
	if err != nil {
		return err
	}
	if len(deletedUsers) != len(UIDs) {
		ids := make([]string, 0, len(UIDs)-len(deletedUsers))
		for _, user := range deletedUsers {
			if !slices.Contains(UIDs, user.ID) {
				ids = append(ids, user.ID)
			}
		}
		var message string
		if newState == user.DeletedState {
			message = "Can't delete already deleted user(-s): " + strings.Join(ids, ", ")
		} else {
			message = "Can't restore non-deleted user(-s): " + strings.Join(ids, ", ")
		}
		return Error.NewStatusError(message, http.StatusConflict)
	}

	cond = util.Ternary(newState == user.DeletedState, "IS", "IS NOT")
	value := util.Ternary(newState == user.DeletedState, "NOW()", "NULL")

	queries := make([]*query.Query, 0, len(deletedUsers)+1)

	updateQuery := query.New(
		`UPDATE "user" SET deleted_at = `+value+`, version = version + 1 WHERE id = ANY($1) and deleted_at `+cond+` NULL`,
		UIDs,
	)

	queries = append(queries, updateQuery)

	for _, deletedUser := range deletedUsers {
		op := util.Ternary(newState == user.DeletedState, audit.DeleteOperation, audit.RestoreOperation)

		auditDTO := newAuditDTO(op, act.ToUserTargeted(deletedUser.ID), deletedUser)

		queries = append(queries, newAuditQuery(&auditDTO))
	}

	if err := transaction.New(queries...).Exec(connection.Primary); err != nil {
		return err
	}

	UIDs, logins := make([]string, len(deletedUsers)), make([]string, len(deletedUsers))

	for i, deletedUser := range deletedUsers {
		UIDs[i] = deletedUser.ID
		logins[i] = deletedUser.Login
		if newState == user.DeletedState {
			dblog.Logger.Trace("Revoking sessions of user "+deletedUser.ID+"...", nil)
			// TODO find more optimal solution, cuz this one cause a lot of DB queries
			err := m.session.RevokeAllUserSessions(act.ToUserTargeted(deletedUser.ID))
			if err != nil && err != Error.StatusNotFound {
				dblog.Logger.Error("Failed to revoke sessions of user "+deletedUser.ID, err.Error(), nil)
			}
			dblog.Logger.Trace("Revoking sessions of user "+deletedUser.ID+": OK", nil)
		}
	}

	cache.BulkInvalidateBasicUserDTO(UIDs, logins)

	return nil
}

func (m *Manager) BulkSoftDelete(act *ActionDTO.Basic, UIDs []string) *Error.Status {
	uidsStr := strings.Join(UIDs, ", ")

	dblog.Logger.Info("Bulk soft deleting users "+uidsStr+"...", nil)

	if err := authz.User.SoftDeleteUser(false, act.RequesterRoles); err != nil {
		return err
	}

	if err := m.bulkStateUpdate(user.DeletedState, act, UIDs); err != nil {
		return err
	}

	dblog.Logger.Info("Bulk soft deleting users "+uidsStr+": OK", nil)

	return nil
}

func (m *Manager) BulkRestore(act *ActionDTO.Basic, UIDs []string) *Error.Status {
	uidsStr := strings.Join(UIDs, ", ")

	dblog.Logger.Info("Bulk soft deleting users "+uidsStr+"...", nil)

	if err := authz.User.RestoreUser(act.RequesterRoles); err != nil {
		return err
	}

	if err := m.bulkStateUpdate(user.NotDeletedState, act, UIDs); err != nil {
		return err
	}

	dblog.Logger.Info("Bulk soft deleting users "+uidsStr+": OK", nil)

	return nil
}

// TODO add audit (there are some problem with foreign keys)
func (m *Manager) Drop(act *ActionDTO.UserTargeted) *Error.Status {
	dblog.Logger.Info("Dropping soft deleted user...", nil)

	if err := act.ValidateUIDs(); err != nil {
		dblog.Logger.Error("Failed to drop soft deleted user", err.Error(), nil)
		return err
	}

	if err := authz.User.DropUser(act.RequesterRoles); err != nil {
		return err
	}

	user, err := m.GetSoftDeletedUserByID(act.TargetUID)
	if err != nil {
		if err != Error.StatusNotFound {
			return err
		}

		user, err = m.GetUserByID(act.TargetUID)
		if err != nil {
			return err
		}

		errMsg := "Only soft deleted users can be dropped"
		dblog.Logger.Error("Failed to drop soft deleted user", errMsg, nil)
		return Error.NewStatusError(errMsg, http.StatusBadRequest)
	}

	deleteQuery := query.New(
		`DELETE FROM "user"
        WHERE id = $1 AND deleted_at IS NOT NULL;`,
		act.TargetUID,
	)
	if err := executor.Exec(connection.Primary, deleteQuery); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = new(time.Time)
	invalidateBasicUserDtoCache(user, updatedUser)

	dblog.Logger.Info("Dropping soft deleted user: OK", nil)

	return nil
}

// TODO add audit (this method really cause a lot of problems)
func (m *Manager) DropAllSoftDeleted(act *ActionDTO.Basic) *Error.Status {
	dblog.Logger.Info("Dropping all soft deleted users...", nil)

	if err := act.ValidateRequesterUID(); err != nil {
		dblog.Logger.Error("Failed to drop all soft deleted users", err.Error(), nil)
		return err
	}

	if err := authz.User.DropAllSoftDeletedUsers(act.RequesterRoles); err != nil {
		return err
	}

	user, err := m.GetUserByID(act.RequesterUID)
	if err != nil {
		return err
	}

	// it's not necessary, but may it be here.
	// Some additional security checks won't be a problem.
	for _, role := range user.Roles {
		if !slices.Contains(act.RequesterRoles, role) {
			errMsg := "Your roles differs on server, try to re-logging in"
			dblog.Logger.Error("Failed to drop all soft deleted users", errMsg, nil)
			return Error.NewStatusError(errMsg, http.StatusConflict)
		}
	}

	deleteQuery := query.New(
		`DELETE FROM "user"
        WHERE deleted_at IS NOT NULL;`,
	)

	err = executor.Exec(connection.Primary, deleteQuery)

	cache.Client.ProgressiveDeletePattern(cache.DeletedUserKeyPrefix + "*")

	dblog.Logger.Info("Dropping all soft deleted users: OK", nil)

	return err
}
