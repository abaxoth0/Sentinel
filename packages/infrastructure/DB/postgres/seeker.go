package postgres

import (
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/cache"
	"strconv"
)

type seeker struct {
    //
}

var invalidUID = Error.NewStatusError("Invalid ID", http.StatusBadRequest)

func (_ *seeker) FindUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    parsedID, e := strconv.ParseInt(id, 10, 64);

    if e != nil {
        return nil, invalidUID
    }

    return queryDTO(
        cache.UserKeyPrefix + "id:" + id,
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE id = $1 AND deletedAt = 0;`,
        parsedID,
    )
}

func (_ *seeker) FindSoftDeletedUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    parsedID, e := strconv.ParseInt(id, 10, 64);

    if e != nil {
        return nil, invalidUID
    }

    return queryDTO(
        cache.DeletedUserKeyPrefix + "id:" + id,
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE id = $1 AND deletedAt <> 0;`,
        parsedID,
    )
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

