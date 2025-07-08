package postgres

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	"slices"
	"strings"
	"time"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type repository struct {
    //
}

func invalidateBasicUserDtoCache(old, current *UserDTO.Basic) {
	invalidator := cache.NewBasicUserDtoInvalidator(old, current)
	if err := invalidator.Invalidate(); err != nil {
		dbLogger.Error("Failed to invalidate cache", err.Error(), nil)
	}
}

func (_ *repository) checkLogin(login string) *Error.Status {
    if err := user.ValidateLogin(login); err != nil {
        return err
    }

    _, err := driver.FindAnyUserByLogin(login)
    if err != nil {
        // user wasn't found, hence login is free to use
        if err.Status() == http.StatusNotFound {
            return nil
        }

        return err
    }

    // if there are no any error (which means that user with this login exists)
    return loginAlreadyInUse
}

func (r *repository) Create(login string, password string) (string, *Error.Status) {
    if err := r.checkLogin(login); err != nil {
        return "", err
    }

    if err := user.ValidatePassword(password); err != nil {
        return "", err
    }

	hashedPassword, err := hashPassword(password)
    if err != nil {
        return "", nil
    }

    uid := uuid.New()

    query := newQuery(
        `INSERT INTO "user" (id, login, password, roles) VALUES
        ($1, $2, $3, $4);`,
        uid, login, hashedPassword, rbac.GetRolesNames(authz.Host.DefaultRoles),
    )

    if err = cache.Client.DeleteOnNoError(
        query.Exec(primaryConnection),
        cache.KeyBase[cache.UserByLogin] + login,
        cache.KeyBase[cache.AnyUserByLogin] + login,
    ); err != nil {
        return "", err
    }

    return uid.String(), nil
}

func (_ *repository) SoftDelete(act *ActionDTO.Targeted) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

	if err := authz.User.SoftDeleteUser(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := driver.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    audit := newAudit(deleteOperation, act, user)

    query := newQuery(
        `UPDATE "user" SET deleted_at = $1, version = version + 1
        WHERE id = $2 AND deleted_at IS NULL;`,
        // deleted_at set manualy instead of using NOW()
        // cuz changed_at and deleted_at should be synchronized
        audit.ChangedAt, act.TargetUID,
    )

    if err := execWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = audit.ChangedAt
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

func (_ *repository) Restore(act *ActionDTO.Targeted) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

	if err := authz.User.RestoreUser(act.RequesterRoles); err != nil {
		return err
	}

    user, err := driver.FindSoftDeletedUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    audit := newAudit(restoreOperation, act, user)

    query := newQuery(
        `UPDATE "user" SET deleted_at = NULL, version = version + 1
        WHERE id = $1 AND deleted_at IS NOT NULL;`,
        act.TargetUID,
    )
    if err := execWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = time.Time{}
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

// TODO add audit
func (_ *repository) bulkStateUpdate(newState user.State, act *ActionDTO.Basic, UIDs []string) *Error.Status {
	if newState != user.DeletedState && newState != user.NotDeletedState {
		dbLogger.Panic(
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

	cond := util.Ternary(newState == user.DeletedState, "IS NOT", "IS")
	deletedUsers, err := newQuery(
		`SELECT id, login, password, roles, deleted_at FROM "user" WHERE id = ANY($1) and deleted_at `+cond+` NULL;`,
		UIDs,
	).CollectBasicUserDTO(primaryConnection)
	if err != Error.StatusNotFound {
		if err != nil  {
			return err
		}
		ids := make([]string, len(deletedUsers))
		for i, user := range deletedUsers {
			ids[i] = user.ID
		}
		return Error.NewStatusError(
			util.Ternary(
				newState == user.DeletedState,
				"Can't delete already deleted user(-s): " + strings.Join(ids, ", "),
				"Can't restore non-deleted user(-s): " + strings.Join(ids, ", "),
			),
			http.StatusConflict,
		)
	}

	cond = util.Ternary(newState == user.DeletedState, "IS", "IS NOT")
	value := util.Ternary(newState == user.DeletedState, "NOW()", "NULL")
	err = newQuery(
		`UPDATE "user" SET deleted_at = `+value+`, version = version + 1 WHERE id = ANY($1) and deleted_at `+cond+` NULL`,
		UIDs,
	).Exec(primaryConnection)
	if err != nil {
		return err
	}

	UIDs, logins := make([]string, len(deletedUsers)), make([]string, len(deletedUsers))

	for i, user := range deletedUsers {
		UIDs[i] = user.ID
		logins[i] = user.Login
	}

	if err := cache.BulkInvalidateBasicUserDTO(UIDs, logins); err != nil {
		dbLogger.Error(
			"Failed to bulk invaldiate users cache",
			err.Error(),
			nil,
		)
	}

	return nil
}

func (_ *repository) BulkSoftDelete(act *ActionDTO.Basic, UIDs []string) *Error.Status {
	if err := authz.User.SoftDeleteUser(false, act.RequesterRoles); err != nil {
		return err
	}
	return driver.bulkStateUpdate(user.DeletedState, act, UIDs)
}

func (_ *repository) BulkRestore(act *ActionDTO.Basic, UIDs []string) *Error.Status {
	if err := authz.User.RestoreUser(act.RequesterRoles); err != nil {
		return err
	}
	return driver.bulkStateUpdate(user.NotDeletedState, act, UIDs)
}

// TODO add audit (there are some problem with foreign keys)
func (_ *repository) Drop(act *ActionDTO.Targeted) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

	if err := authz.User.DropUser(act.RequesterRoles); err != nil {
		return err
	}

    user, err := driver.FindAnyUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    if user.DeletedAt.IsZero() {
        return Error.NewStatusError(
            "Only soft deleted users can be dropped",
            http.StatusBadRequest,
        )
    }

    query := newQuery(
        `DELETE FROM "user"
        WHERE id = $1 AND deleted_at IS NOT NULL;`,
        act.TargetUID,
    )
    if err := query.Exec(primaryConnection); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.DeletedAt = time.Time{}
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

// TODO add audit (this method really cause a lot of problems)
func (_ *repository) DropAllSoftDeleted(act *ActionDTO.Basic) *Error.Status {
    if err := act.ValidateRequesterUID(); err != nil {
        return err
    }

	if err := authz.User.DropAllSoftDeletedUsers(act.RequesterRoles); err != nil {
		return err
	}

    user, err := driver.FindUserByID(act.RequesterUID)
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

    query := newQuery(
        `DELETE FROM "user"
        WHERE deleted_at IS NOT NULL;`,
    )

    err = query.Exec(primaryConnection)

    cache.Client.ProgressiveDeletePattern(cache.DeletedUserKeyPrefix + "*")

    return err
}

func (r *repository) ChangeLogin(act *ActionDTO.Targeted, newLogin string) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

    if err := user.ValidateLogin(newLogin); err != nil {
        return err
    }

	if err := authz.User.ChangeUserLogin(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := driver.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    if user.Login == newLogin {
        return Error.NewStatusError(
            "Login not changed: Current login and new login are the same",
            http.StatusConflict,
        )
    }

    if err := r.checkLogin(newLogin); err != nil {
        return err
    }

    audit := newAudit(updatedOperation, act, user)

    query := newQuery(
        `UPDATE "user" SET login = $1, version = version + 1
        WHERE id = $2;`,
        newLogin, act.TargetUID,
    )

    if err := execWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.Login = newLogin
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

func (_ *repository) ChangePassword(act *ActionDTO.Targeted, newPassword string) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

    if err := user.ValidatePassword(newPassword); err != nil {
        return err
    }

	if err := authz.User.ChangeUserPassword(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := driver.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

	hashedPassword, e := hashPassword(newPassword)
    if e != nil {
        return e
    }

    audit := newAudit(updatedOperation, act, user)

    query := newQuery(
        `UPDATE "user" SET password = $1, version = version + 1
        WHERE id = $2;`,
        hashedPassword, act.TargetUID,
    )

    if err := execWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.Password = string(hashedPassword)
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

func (_ *repository) ChangeRoles(act *ActionDTO.Targeted, newRoles []string) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

    if act.TargetUID == act.RequesterUID &&
       slices.Contains(act.RequesterRoles, "admin") &&
       !slices.Contains(newRoles, "admin") {
          return Error.NewStatusError(
              "Нельзя снять роль администратора с самого себя",
               http.StatusForbidden,
          )
    }

    user, err := driver.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

	if err := authz.User.ChangeUserRoles(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    audit := newAudit(updatedOperation, act, user)

    query := newQuery(
        `UPDATE "user" SET roles = $1, version = version + 1
        WHERE id = $2;`,
        newRoles, act.TargetUID,
    )

    if err := execWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.Roles = newRoles
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

func (_ *repository) Activate(tk string) *Error.Status {
    t, err := token.ParseSingedToken(tk, config.Secret.ActivationTokenPublicKey)
    if err != nil {
        return err
    }

    payload, err := UserMapper.PayloadFromClaims(t.Claims.(jwt.MapClaims))
    if err != nil {
        return err
    }

    user, err := driver.FindUserByID(payload.ID)
    if err != nil {
        return err
    }
    if user.IsActive() {
        return Error.NewStatusError(
            "User already active",
            http.StatusConflict,
        )
    }

    filter := ActionDTO.NewTargeted(user.ID, user.ID, user.Roles)

    audit := newAudit(updatedOperation, filter, user)
	var updatedUser *UserDTO.Basic

    for i, role := range user.Roles {
        if role == "unconfirmed_user" {
			updatedUser = user.Copy()
            updatedUser.Roles[i] = "user"
			updatedUser.Version++
            break
        }
    }

	// This should be impossible, but additional check won't be redundant
	if updatedUser == nil {
		dbLogger.Error(
			"Failed to activate user " + user.ID,
			"User doesn't have role unconfirmed_user: " + strings.Join(user.Roles, ","),
			nil,
		)
		return Error.StatusInternalError
	}

    tx := newTransaction(
        newQuery(
            `UPDATE "user" SET roles = $1, version = version + 1
             WHERE login = $2;`,
             updatedUser.Roles, updatedUser.Login,
        ),
        newAuditQuery(&audit),
    )
    if err := tx.Exec(); err != nil {
        return err
    }

	invalidateBasicUserDtoCache(user, updatedUser)

    return nil
}

