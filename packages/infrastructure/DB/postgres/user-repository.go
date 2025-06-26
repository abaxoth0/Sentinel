package postgres

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	"slices"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type repository struct {
    //
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
        // TODO try to replace that everywhere with cache.client.DeletePattern
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
        `UPDATE "user" SET deleted_at = $1
        WHERE id = $2 AND deleted_at IS NULL;`,
        // deleted_at set manualy instead of using NOW()
        // cuz changed_at and deleted_at should be synchronized
        audit.ChangedAt, act.TargetUID,
    )

    return cache.Client.DeleteOnNoError(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + act.TargetUID,
        cache.KeyBase[cache.DeletedUserById] + act.TargetUID,
        cache.KeyBase[cache.AnyUserById] + act.TargetUID,
        cache.KeyBase[cache.UserRolesById] + act.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
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
        `UPDATE "user" SET deleted_at = NULL
        WHERE id = $1 AND deleted_at IS NOT NULL;`,
        act.TargetUID,
    )

    return cache.Client.DeleteOnNoError(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + act.TargetUID,
        cache.KeyBase[cache.DeletedUserById] + act.TargetUID,
        cache.KeyBase[cache.AnyUserById] + act.TargetUID,
        cache.KeyBase[cache.UserRolesById] + act.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
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

    return cache.Client.DeleteOnNoError(
        query.Exec(primaryConnection),
        cache.KeyBase[cache.DeletedUserById] + act.TargetUID,
        cache.KeyBase[cache.AnyUserById] + act.TargetUID,
        cache.KeyBase[cache.UserRolesById] + act.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
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
        `UPDATE "user" SET login = $1
        WHERE id = $2;`,
        newLogin, act.TargetUID,
    )

    return cache.Client.DeleteOnNoError(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + act.TargetUID,
        cache.KeyBase[cache.AnyUserById] + act.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
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
        `UPDATE "user" SET password = $1
        WHERE id = $2;`,
        hashedPassword, act.TargetUID,
    )

    return cache.Client.DeleteOnNoError(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + act.TargetUID,
        cache.KeyBase[cache.AnyUserById] + act.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
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
        `UPDATE "user" SET roles = $1
        WHERE id = $2;`,
        newRoles, act.TargetUID,
    )

    return cache.Client.DeleteOnNoError(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + act.TargetUID,
        cache.KeyBase[cache.AnyUserById] + act.TargetUID,
        cache.KeyBase[cache.UserRolesById] + act.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
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

    for i, role := range user.Roles {
        if role == "unconfirmed_user" {
            user.Roles[i] = "user"
            break
        }
    }

    tx := newTransaction(
        newQuery(
            `UPDATE "user" SET roles = $1
             WHERE login = $2;`,
             user.Roles, user.Login,
        ),
        newAuditQuery(&audit),
    )
    if err := tx.Exec(); err != nil {
        return err
    }

    if err := cache.Client.Delete(
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    ); err != nil {
        return Error.StatusInternalError
    }

    return nil
}

