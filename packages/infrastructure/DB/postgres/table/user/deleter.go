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
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	SessionTable "sentinel/packages/infrastructure/DB/postgres/table/session"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	"slices"
	"strings"
	"time"
)

func (m *Manager) SoftDelete(act *ActionDTO.UserTargeted) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

	if err := authz.User.SoftDeleteUser(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := m.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    auditUserDTO := audit.NewUser(audit.DeleteOperation, act, user)

    updateQuery := query.New(
        `UPDATE "user" SET deleted_at = $1, version = version + 1
        WHERE id = $2 AND deleted_at IS NULL;`,
        // deleted_at set manualy instead of using NOW()
        // cuz changed_at and deleted_at should be synchronized
        auditUserDTO.ChangedAt, act.TargetUID,
    )
	sessionsDeleteQuery := SessionTable.NewRevokeAllUserSessionsQuery(act)

    if err := audit.ExecTxWithAuditUser(&auditUserDTO, updateQuery, sessionsDeleteQuery); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = auditUserDTO.ChangedAt
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	if err := m.session.DeleteUserSessionsCache(user.ID); err != nil {
		userLogger.Error("Failed to delete user sessions cache", err.Error(), nil)
	}

	return nil
}

func (m *Manager) Restore(act *ActionDTO.UserTargeted) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

	if err := authz.User.RestoreUser(act.RequesterRoles); err != nil {
		return err
	}

    user, err := m.FindSoftDeletedUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    auditUserDTO := audit.NewUser(audit.RestoreOperation, act, user)

    query := query.New(
        `UPDATE "user" SET deleted_at = NULL, version = version + 1
        WHERE id = $1 AND deleted_at IS NOT NULL;`,
        act.TargetUID,
    )
    if err := audit.ExecTxWithAuditUser(&auditUserDTO, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = time.Time{}
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

// TODO add auditUserDTO
func (m *Manager) bulkStateUpdate(newState user.State, act *ActionDTO.Basic, UIDs []string) *Error.Status {
	if newState != user.DeletedState && newState != user.NotDeletedState {
		userLogger.Panic(
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
				"Invalid user ID format (expected UUID): " + uid,
			)
		}
	}

	cond := util.Ternary(newState == user.DeletedState, "IS", "IS NOT")
	selectQuery := query.New(
		`SELECT id, login, password, roles, deleted_at, version FROM "user" WHERE id = ANY($1) and deleted_at `+cond+` NULL;`,
		UIDs,
	)

	deletedUsers, err := executor.CollectBasicUserDTO(connection.Replica, selectQuery)
	if err != nil {
		return err
	}
	if len(deletedUsers) != len(UIDs) {
		ids := make([]string, 0, len(UIDs) - len(deletedUsers))
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
	updateQuery := query.New(
		`UPDATE "user" SET deleted_at = `+value+`, version = version + 1 WHERE id = ANY($1) and deleted_at `+cond+` NULL`,
		UIDs,
	)

	err = executor.Exec(connection.Primary, updateQuery)
	if err != nil {
		return err
	}

	UIDs, logins := make([]string, len(deletedUsers)), make([]string, len(deletedUsers))

	for i, deletedUser := range deletedUsers {
		UIDs[i] = deletedUser.ID
		logins[i] = deletedUser.Login
		if newState == user.DeletedState {
			userLogger.Trace("Revoking sessions of user "+deletedUser.ID+"...", nil)
			// TODO find more optimal solution, cuz this one cause a lot of DB queries
			err := m.session.RevokeAllUserSessions(act.ToUserTargeted(deletedUser.ID))
			if err != nil && err != Error.StatusNotFound {
				userLogger.Error("Failed to revoke sessions of user "+deletedUser.ID, err.Error(), nil)
			}
			userLogger.Trace("Revoking sessions of user "+deletedUser.ID+": OK", nil)
		}
	}

	if err := cache.BulkInvalidateBasicUserDTO(UIDs, logins); err != nil {
		userLogger.Error(
			"Failed to bulk invaldiate users cache",
			err.Error(),
			nil,
		)
	}

	return nil
}

func (m *Manager) BulkSoftDelete(act *ActionDTO.Basic, UIDs []string) *Error.Status {
	if err := authz.User.SoftDeleteUser(false, act.RequesterRoles); err != nil {
		return err
	}
	if err := m.bulkStateUpdate(user.DeletedState, act, UIDs); err != nil {
		return err
	}

	return nil
}

func (m *Manager) BulkRestore(act *ActionDTO.Basic, UIDs []string) *Error.Status {
	if err := authz.User.RestoreUser(act.RequesterRoles); err != nil {
		return err
	}
	return m.bulkStateUpdate(user.NotDeletedState, act, UIDs)
}

// TODO add auditUserDTO (there are some problem with foreign keys)
func (m *Manager) Drop(act *ActionDTO.UserTargeted) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

	if err := authz.User.DropUser(act.RequesterRoles); err != nil {
		return err
	}

    user, err := m.FindAnyUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    if user.DeletedAt.IsZero() {
        return Error.NewStatusError(
            "Only soft deleted users can be dropped",
            http.StatusBadRequest,
        )
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
	updatedUser.DeletedAt = time.Time{}
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

// TODO add auditUserDTO (this method really cause a lot of problems)
func (m *Manager) DropAllSoftDeleted(act *ActionDTO.Basic) *Error.Status {
    if err := act.ValidateRequesterUID(); err != nil {
        return err
    }

	if err := authz.User.DropAllSoftDeletedUsers(act.RequesterRoles); err != nil {
		return err
	}

    user, err := m.FindUserByID(act.RequesterUID)
    if err != nil {
        return err
    }

    // it's not necessary, but may it be here.
    // Some additional security checks won't be a problem.
    for _, role := range user.Roles {
        if !slices.Contains(act.RequesterRoles, role) {
            return Error.NewStatusError(
                "Your roles differs on server, try to re-logging in",
                http.StatusConflict,
            )
        }
    }

    deleteQuery := query.New(
        `DELETE FROM "user"
        WHERE deleted_at IS NOT NULL;`,
    )

    err = executor.Exec(connection.Primary, deleteQuery)

    cache.Client.ProgressiveDeletePattern(cache.DeletedUserKeyPrefix + "*")

    return err
}

