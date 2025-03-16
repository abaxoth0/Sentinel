package postgres

import (
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"strings"
)

type seeker struct {
    //
}

// TODO now value can be only a string, rework that.
//      (make it a function with some generic instead of seeker's method?)
func (s *seeker) findUserBy(
    property userProperty,
    propertyValue string,
    state userState,
    cacheKey string,
) (*UserDTO.Indexed, *Error.Status) {
    user, err := queryDTO(
        cacheKey,
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE ` + string(property) + ` = $1;`,
        propertyValue,
    )

    if err != nil {
        return nil, err
    }

    if state == deletedUserState && user.DeletedAt == 0 {
        return nil, Error.StatusUserNotFound
    }

    if state == notDeletedUserState && user.DeletedAt != 0 {
        return nil, Error.StatusUserNotFound
    }

    return user, nil
}

func (s *seeker) FindAnyUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    return s.findUserBy(idProperty, id, anyUserState, anyUserCacheKeyBase + id)
}

func (s *seeker) FindUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    return s.findUserBy(idProperty, id, notDeletedUserState, userCacheKeyBase + id)
}

func (s *seeker) FindSoftDeletedUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    return s.findUserBy(idProperty, id, deletedUserState, deletedUserCacheKeyBase + id)
}

func (s *seeker) FindUserByLogin(login string) (*UserDTO.Indexed, *Error.Status) {
    return s.findUserBy(loginProperty, login, anyUserState, cache.UserKeyPrefix + "any_login:" + login)
}

func (_ *seeker) IsLoginExists(target string) (bool, *Error.Status) {
    cacheKey := cache.UserKeyPrefix + "loginExists:" + target

    if cachedData, hit := cache.Client.Get(cacheKey); hit {
        return cachedData == "true", nil
    }

    var id int

    scan, err := queryRow(
        `SELECT id
         FROM "user"
         WHERE login = $1 and deletedAt = 0;`,
        target,
    )

    if err != nil {
        return false, err
    }

    if e := scan(false, &id); e != nil {
        if e == Error.StatusUserNotFound {
            cache.Client.Set(cacheKey, false)
            return false, nil
        }

        return false, e
    }

    cache.Client.Set(cacheKey, true)

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

    if rawRoles, hit := cache.Client.Get(userRolesCacheKeyBase + filter.TargetUID); hit {
        return strings.Split(rawRoles, ","), nil
    }

    sql := `SELECT roles FROM "user" WHERE id = $1 AND deletedAt = 0;`

    scan, err := queryRow(sql, filter.TargetUID)

    if err != nil {
        return nil, err
    }

    roles := []string{}

    if e := scan(false, &roles); e != nil {
        return nil, e
    }

    cache.Client.Set(userRolesCacheKeyBase + filter.TargetUID, strings.Join(roles, ","))

    return roles, nil
}

