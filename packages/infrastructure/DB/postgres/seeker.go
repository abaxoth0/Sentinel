package postgres

import (
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"strconv"
	"strings"
)

type seeker struct {
    //
}

var invalidUID = Error.NewStatusError("Invalid ID", http.StatusBadRequest)

func (_ *seeker) findAnyUserByID(id string, cacheKey string) (*UserDTO.Indexed, *Error.Status) {
    parsedID, e := strconv.ParseInt(id, 10, 64);

    if e != nil {
        return nil, invalidUID
    }

    return queryDTO(
        cacheKey,
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE id = $1;`,
        parsedID,
    )
}


func (s *seeker) findUserByID(id string, deleted bool) (*UserDTO.Indexed, *Error.Status) {
    var cacheKey string

    if deleted {
        cacheKey = deletedUserCacheKeyBase + id
    } else {
        cacheKey = userCacheKeyBase + id
    }

    user, err := s.findAnyUserByID(id, cacheKey)

    if err != nil {
        return nil, err
    }

    if deleted && user.DeletedAt == 0 {
        return nil, Error.StatusUserNotFound
    }

    return user, nil
}

func (s *seeker) FindAnyUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    // TODO invalidate
    return s.findAnyUserByID(id, anyUserCacheKeyBase + id)
}

func (s *seeker) FindUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    return s.findUserByID(id, false)
}

func (s *seeker) FindSoftDeletedUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    return s.findUserByID(id, true)
}

func (_ *seeker) FindUserByLogin(login string) (*UserDTO.Indexed, *Error.Status) {
    return queryDTO(
        cache.UserKeyPrefix + "login:" + login,
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE login = $1 AND deletedAt = 0;`,
        login,
    )
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

