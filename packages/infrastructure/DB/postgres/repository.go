package postgres

import (
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/util"
	"slices"
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

// TODO Check if user exists
func (_ *repository) Create(login string, password string) (*Error.Status) {
	hashedPassword, err := hashPassword(password)

    if err != nil {
        return nil
    }

    return queryExec(
        `INSERT INTO "user" (login, password, roles, deletedAt) VALUES
        ($1, $2, $3, $4);`,
        login, hashedPassword, []string{authorization.Host.OriginRoleName}, 0,
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
            `UPDATE "user" SET deletedAt = $1
             WHERE id = $2 AND deletedAt = 0;`,
             util.UnixTimeNow(), filter.TargetUID,
        ),
        userCacheKey(filter.TargetUID, false),
        userRolesCacheKey(filter.TargetUID),
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

    _, err := driver.FindSoftDeletedUserByID(filter.TargetUID)

    if err != nil {
        return err
    }

    return handleUserCache(
        queryExec(
            `UPDATE "user" SET deletedAt = 0
             WHERE id = $1 AND deletedAt <> 0;`,
             filter.TargetUID,
        ),
        userCacheKey(filter.TargetUID, true),
        userRolesCacheKey(filter.TargetUID),
    )
}

// TODO Allow to hard deleted only soft deleted users
func (_ *repository) Drop(filter *UserDTO.Filter) *Error.Status {
    if err := authorization.Authorize(
        authorization.Action.Drop,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    return handleUserCache(
        queryExec(
            `DELETE FROM "user"
             WHERE id = $1;`,
             filter.TargetUID,
        ),
        userCacheKey(filter.TargetUID, true),
        userRolesCacheKey(filter.TargetUID),
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
             WHERE deletedAt <> 0;`,
        ),
        userCacheKey(filter.TargetUID, true),
        userRolesCacheKey(filter.TargetUID),
    )
}

func (_ *repository) ChangeLogin(filter *UserDTO.Filter, newLogin string) *Error.Status {
    if err := authorization.Authorize(
        authorization.Action.ChangeLogin,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return err
    }

    return handleUserCache(
        queryExec(
            `UPDATE "user" SET login = $1
             WHERE id = $2;`,
             newLogin, filter.TargetUID,
        ),
        userCacheKey(filter.TargetUID, false),
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
        userCacheKey(filter.TargetUID, false),
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
        userCacheKey(filter.TargetUID, false),
        userRolesCacheKey(filter.TargetUID),
    )
}

