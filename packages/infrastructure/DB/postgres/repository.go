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

    return handleUserCache(
        queryExec(
            `INSERT INTO "user" (id, login, password, roles, deleted_at) VALUES
             ($1, $2, $3, $4, $5);`,
             uuid.New(), login, hashedPassword, []string{authorization.Host.OriginRoleName}, nil,
        ),
        cache.KeyBase[cache.UserByLogin] + login,
        cache.KeyBase[cache.AnyUserByLogin] + login,
    )
}

// TODO Create new table for soft deleted users
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

    return handleUserCache(
        queryExec(
            `UPDATE "user" SET deleted_at = NOW()
             WHERE id = $1 AND deleted_at IS NULL;`,
             filter.TargetUID,
        ),
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

    return handleUserCache(
        queryExec(
            `UPDATE "user" SET deleted_at = NULL
             WHERE id = $1 AND deleted_at IS NOT NULL;`,
             filter.TargetUID,
        ),
        // TODO ... is that just me or it's looks kinda bad?
        //      Try to find a better way to invalidate cache
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.DeletedUserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserRolesById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

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

    return handleUserCache(
        queryExec(
            `DELETE FROM "user"
             WHERE id = $1 AND deleted_at IS NOT NULL;`,
             filter.TargetUID,
        ),
        cache.KeyBase[cache.DeletedUserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserRolesById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

func (_ *repository) DropAllSoftDeleted(filter *UserDTO.Filter) *Error.Status {
    if err := authorization.Authorize(
        authorization.Action.DropAllSoftDeleted,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    return handleUserCache(
        queryExec(
            `DELETE FROM "user"
             WHERE deleted_at IS NOT NULL;`,
        ),
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

    return handleUserCache(
        queryExec(
            `UPDATE "user" SET login = $1
             WHERE id = $2;`,
             newLogin, filter.TargetUID,
        ),
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

    return handleUserCache(
        queryExec(
            `UPDATE "user" SET password = $1
             WHERE id = $2;`,
            hashedPassword, filter.TargetUID,
        ),
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

    return handleUserCache(
        queryExec(
            `UPDATE "user" SET roles = $1
             WHERE id = $2;`,
             newRoles, filter.TargetUID,
        ),
        cache.KeyBase[cache.UserById] + filter.TargetUID,
        cache.KeyBase[cache.AnyUserById] + filter.TargetUID,
        cache.KeyBase[cache.UserRolesById] + filter.TargetUID,
        cache.KeyBase[cache.UserByLogin] + user.Login,
        cache.KeyBase[cache.AnyUserByLogin] + user.Login,
    )
}

