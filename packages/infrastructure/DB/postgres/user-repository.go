package postgres

import (
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"slices"

	"github.com/google/uuid"
)

type repository struct {
    //
}

func handleUserCache(err *Error.Status, keys ...string) *Error.Status {
    if err == nil {
        for _, key := range keys {
            cache.Client.Delete(key)
        }
    }

    return err
}

var loginAlreadyInUse = Error.NewStatusError(
    "Login already in use",
    http.StatusConflict,
)

func (_ *repository) checkLogin(login string) *Error.Status {
    _, err := driver.FindAnyUserByLogin(login)

    if err != Error.StatusUserNotFound {
        // if error is persist, but it's not an Error.StatusUserNotFound
        if err != nil {
            return err
        }

        // if there are no any error (which means that user with this login exists)
        return loginAlreadyInUse
    }

    // login is free to use, there are no error
    return nil
}


func (r *repository) Create(login string, password string) (*Error.Status) {
    if err := r.checkLogin(login); err != nil {
        return err
    }

	hashedPassword, err := hashPassword(password)

    if err != nil {
        return nil
    }

    query := newQuery(
        `INSERT INTO "user" (id, login, password, roles) VALUES
        ($1, $2, $3, $4);`,
        uuid.New(), login, hashedPassword, []string{authorization.Host.OriginRoleName},
    )

    return handleUserCache(
        query.Exec(),
        cache.KeyBase[cache.UserByLogin] + login,
        cache.KeyBase[cache.AnyUserByLogin] + login,
    )
}

func (_ *repository) SoftDelete(filter *UserDTO.Filter) *Error.Status {
    // TODO add possibility to config what kind of users can delete themselves
    // all users can delete themselves, except admins (TEMP)
    if filter.TargetUID != filter.RequesterUID {
        if err := authorization.Authorize(
            authorization.Action.SoftDelete,
            authorization.Resource.User,
            filter.RequesterRoles,
        ); err != nil {
            return err
        }
    }

    user, err := driver.FindUserByID(filter.TargetUID)

    if err != nil {
        return err
    }

    if slices.Contains(user.Roles, "admin") {
        return Error.NewStatusError(
            "Нельзя удалить пользователя с ролью администратора",
            http.StatusBadRequest,
        )
    }

    audit := newAudit(deleteOperation, filter, user)

    query := newQuery(
        `UPDATE "user" SET deleted_at = $1
        WHERE id = $2 AND deleted_at IS NULL;`,
        // deleted_at set manualy instead of using NOW()
        // cuz changed_at and deleted_at should be synchronized
        audit.ChangedAt, filter.TargetUID,
    )

    return handleUserCache(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.DeletedUserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserRolesById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

func (_ *repository) Restore(filter *UserDTO.Filter) *Error.Status {
    if err := authorization.Authorize(
        authorization.Action.Restore,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    user, err := driver.FindSoftDeletedUserByID(filter.TargetUID)

    if err != nil {
        return err
    }

    audit := newAudit(restoreOperation, filter, user)

    query := newQuery(
        `UPDATE "user" SET deleted_at = NULL
        WHERE id = $1 AND deleted_at IS NOT NULL;`,
        filter.TargetUID,
    )

    return handleUserCache(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.DeletedUserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserRolesById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

// TODO add audit (there are some problem with foreign keys)
func (_ *repository) Drop(filter *UserDTO.Filter) *Error.Status {
    if err := authorization.Authorize(
        authorization.Action.Drop,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    user, err := driver.FindAnyUserByID(filter.TargetUID)

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
        filter.TargetUID,
    )

    return handleUserCache(
        query.Exec(),
        cache.KeyBase[cache.DeletedUserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserRolesById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

// TODO add audit (this method really cause a lot of problems)
func (_ *repository) DropAllSoftDeleted(filter *UserDTO.Filter) *Error.Status {
    if err := authorization.Authorize(
        authorization.Action.DropAllSoftDeleted,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    query := newQuery(
        `DELETE FROM "user"
        WHERE deleted_at IS NOT NULL;`,
    )

    return handleUserCache(
        query.Exec(),
        // TODO there are a problem with cache invalidation in this case,
        //      must be deleted all cache for users with 'deleted' and 'any' state,
        //      maybe there are some delete pattern option?
        cache.KeyBase[cache.DeletedUserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
    )
}

func (r *repository) ChangeLogin(filter *UserDTO.Filter, newLogin string) *Error.Status {
    if err := authorization.Authorize(
        authorization.Action.ChangeLogin,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    user, err := driver.FindUserByID(filter.TargetUID)

    if err != nil {
        return err
    }

    if err := r.checkLogin(newLogin); err != nil {
        return err
    }

    audit := newAudit(updatedOperation, filter, user)

    query := newQuery(
        `UPDATE "user" SET login = $1
        WHERE id = $2;`,
        newLogin, filter.TargetUID,
    )

    return handleUserCache(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

func (_ *repository) ChangePassword(filter *UserDTO.Filter, newPassword string) *Error.Status {
    if filter.RequesterUID != filter.TargetUID {
        if err := authorization.Authorize(
            authorization.Action.ChangePassword,
            authorization.Resource.User,
            filter.RequesterRoles,
        ); err != nil {
            return err
        }
    }

    user, err := driver.FindUserByID(filter.TargetUID)

    if err != nil {
        return err
    }

	hashedPassword, e := hashPassword(newPassword)

    if e != nil {
        return e
    }

    audit := newAudit(updatedOperation, filter, user)

    query := newQuery(
        `UPDATE "user" SET password = $1
        WHERE id = $2;`,
        hashedPassword, filter.TargetUID,
    )

    return handleUserCache(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

func (_ *repository) ChangeRoles(filter *UserDTO.Filter, newRoles []string) *Error.Status {
    if filter.TargetUID == filter.RequesterUID &&
       slices.Contains(filter.RequesterRoles, "admin") &&
       !slices.Contains(newRoles, "admin") {
          return Error.NewStatusError(
              "Нельзя снять роль администратора с самого себя",
               http.StatusForbidden,
          )
    }

    user, err := driver.FindUserByID(filter.TargetUID)

    if err != nil {
        return err
    }

    if err := authorization.Authorize(
        authorization.Action.ChangeRoles,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    audit := newAudit(updatedOperation, filter, user)

    query := newQuery(
        `UPDATE "user" SET roles = $1
        WHERE id = $2;`,
        newRoles, filter.TargetUID,
    )

    return handleUserCache(
        execWithAudit(&audit, query),
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserRolesById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

