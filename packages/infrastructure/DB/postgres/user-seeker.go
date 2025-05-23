package postgres

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	actiondto "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	"strings"
)

type seeker struct {
    //
}

// TODO Now users can be found only by one property,
//      there are no way to create some complex search filter.
//      do smth with that.

// TODO now value can be only a string, rework that.
//      (make it a function with some generic instead of seeker's method?)
func (s *seeker) findUserBy(
    conditionProperty user.Property,
    conditionValue string,
    state user.State,
    cacheKey string,
) (*UserDTO.Basic, *Error.Status) {
    if conditionProperty == user.IdProperty {
        if err := validation.UUID(conditionValue); err != nil {
            return nil, err.ToStatus(
                "user id isn't specified",
                "user id has invalid value",
            )
        }
    }
    if conditionProperty == user.LoginProperty && config.App.IsLoginEmail {
        if err := validation.Email(conditionValue); err != nil {
            return nil, err.ToStatus(
                "User login isn't specified",
                "User login has invalid value",
            )
        }
    }

    query := newQuery(
        `SELECT id, login, password, roles, deleted_at
        FROM "user"
        WHERE ` + string(conditionProperty) + ` = $1;`,
        conditionValue,
    )

    dto, err := query.RowBasicUserDTO(cacheKey)

    if err != nil {
        return nil, err
    }

    if state == user.NotDeletedState && dto.IsDeleted() {
        return nil, userNotFound
    }

    if state == user.DeletedState && !dto.IsDeleted() {
        return nil, userNotFound
    }

    return dto, nil
}

func (s *seeker) FindAnyUserByID(id string) (*UserDTO.Basic, *Error.Status) {
    return s.findUserBy(
        user.IdProperty,
        id,
        user.AnyState,
        cache.KeyBase[cache.AnyUserById] + id,
    )
}

func (s *seeker) FindUserByID(id string) (*UserDTO.Basic, *Error.Status) {
    return s.findUserBy(
        user.IdProperty,
        id,
        user.NotDeletedState,
        cache.KeyBase[cache.UserById] + id,
    )
}

func (s *seeker) FindSoftDeletedUserByID(id string) (*UserDTO.Basic, *Error.Status) {
    return s.findUserBy(
        user.IdProperty,
        id,
        user.DeletedState,
        cache.KeyBase[cache.DeletedUserById] + id,
    )
}

func (s *seeker) FindAnyUserByLogin(login string) (*UserDTO.Basic, *Error.Status) {
    return s.findUserBy(
        user.LoginProperty,
        login,
        user.AnyState,
        cache.KeyBase[cache.AnyUserByLogin] + login,
    )
}

func (s *seeker) FindUserByLogin(login string) (*UserDTO.Basic, *Error.Status) {
    return s.findUserBy(
        user.LoginProperty,
        login,
        user.NotDeletedState,
        cache.KeyBase[cache.UserByLogin] + login,
    )
}

func (s *seeker) IsLoginAvailable(login string) (bool, *Error.Status) {
    if err := user.ValidateLogin(login); err != nil {
        return false, err
    }

    cacheKey := cache.KeyBase[cache.UserByLogin] + login

    _, err := s.findUserBy(user.LoginProperty, login, user.NotDeletedState, cacheKey)

    if err != nil {
        if err.Status() == http.StatusNotFound {
            cache.Client.Set(cacheKey, false)
            return false, nil
        }

        return false, err
    }

    return true, nil
}

func (_ *seeker) GetRoles(act *actiondto.Targeted) ([]string, *Error.Status) {
    if err := act.ValidateTargetUID(); err != nil {
        return nil, err
    }

    if err := authz.Authorize(
        authz.Action.GetRoles,
        authz.Resource.User,
        act.RequesterRoles,
    ); err != nil {
        return nil, err
    }

    if rawRoles, hit := cache.Client.Get(cache.KeyBase[cache.UserRolesById] + act.TargetUID); hit {
        return strings.Split(rawRoles, ","), nil
    }

    // TODO is there a point doing that? why just not use DB.Database.FindUserByID()?
    query := newQuery(
        `SELECT roles FROM "user" WHERE id = $1 AND deleted_at IS NULL;`,
        act.TargetUID,
    )

    scan, err := query.Row()

    if err != nil {
        return nil, err
    }

    roles := []string{}

    if e := scan(&roles); e != nil {
        if e == Error.StatusNotFound {
            return nil, userNotFound
        }
        return nil, e
    }

    cache.Client.Set(cache.KeyBase[cache.UserRolesById] + act.TargetUID, strings.Join(roles, ","))

    return roles, nil
}

