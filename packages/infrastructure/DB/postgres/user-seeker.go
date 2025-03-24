package postgres

import (
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/auth/authorization"
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
    query := newQuery(
        `SELECT id, login, password, roles, deleted_at, is_active
        FROM "user"
        WHERE ` + string(conditionProperty) + ` = $1;`,
        conditionValue,
    )

    dto, err := query.RowBasicUserDTO(cacheKey)

    if err != nil {
        return nil, err
    }

    if state == user.NotDeletedState && dto.IsDeleted() {
        return nil, Error.StatusUserNotFound
    }

    if state == user.DeletedState && !dto.IsDeleted() {
        return nil, Error.StatusUserNotFound
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

func (s *seeker) IsLoginExists(login string) (bool, *Error.Status) {
    cacheKey := cache.KeyBase[cache.UserByLogin] + login

    _, err := s.findUserBy(user.LoginProperty, login, user.NotDeletedState, cacheKey)

    if err != nil {
        if err == Error.StatusUserNotFound {
            cache.Client.Set(cacheKey, false)
            return false, nil
        }

        return false, err
    }

    return true, nil
}

func (_ *seeker) GetRoles(filter *UserDTO.Filter) ([]string, *Error.Status) {
    if err := authorization.Authorize(
        authorization.Action.GetRoles,
        authorization.Resource.User,
        filter.RequesterRoles,
    ); err != nil {
        return nil, err
    }

    if rawRoles, hit := cache.Client.Get(cache.KeyBase[cache.UserRolesById] + filter.TargetUID); hit {
        return strings.Split(rawRoles, ","), nil
    }

    query := newQuery(
        `SELECT roles FROM "user" WHERE id = $1 AND deleted_at IS NULL;`,
        filter.TargetUID,
    )

    scan, err := query.Row()

    if err != nil {
        return nil, err
    }

    roles := []string{}

    if e := scan(false, &roles); e != nil {
        return nil, e
    }

    cache.Client.Set(cache.KeyBase[cache.UserRolesById] + filter.TargetUID, strings.Join(roles, ","))

    return roles, nil
}

