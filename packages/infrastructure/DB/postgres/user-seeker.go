package postgres

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserFilterParser "sentinel/packages/infrastructure/parsers/user-filter"
	"strconv"
	"strings"
)

type seeker struct {
    //
}

const (
	searchUsersSqlStart=`WITH numbered_users AS (SELECT *, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) as row_num FROM "user"`
	searchUsersSqlSelect=`SELECT id, login, roles, deleted_at, version FROM numbered_users WHERE row_num BETWEEN `
)

func searchUsersSqlEnd(page, pageSize int) string {
	start := ((page - 1) * pageSize) + 1
	end := page * pageSize
	return ") " + searchUsersSqlSelect + strconv.Itoa(start) + " AND " + strconv.Itoa(end) + ";"
}

func (s *seeker) SearchUsers(
	act *ActionDTO.Basic,
	rawFilters []string,
	page int,
	pageSize int,
) ([]*UserDTO.Public, *Error.Status) {
	if err := authz.User.SearchUsers(act.RequesterRoles); err != nil {
		return nil, err
	}

	if page < 1 {
		return nil, Error.NewStatusError(
			"Invalid page size: "+strconv.Itoa(page)+". It must greater greater than 0.",
			http.StatusBadRequest,
		)
	}
	if pageSize < 1 || pageSize > config.DB.MaxSearchPageSize {
		return nil, Error.NewStatusError(
			"Invalid page size: "+strconv.Itoa(pageSize)+". It must be between 1 and " + strconv.Itoa(config.DB.MaxSearchPageSize),
			http.StatusBadRequest,
		)
	}

	if rawFilters == nil || len(rawFilters) == 0 {
		return nil, Error.NewStatusError(
			"Filter is missing or has invalid format",
			http.StatusBadRequest,
		)
	}

	if len(rawFilters) == 1 && rawFilters[0] == "null" {
		dtos, err := newQuery(searchUsersSqlStart + searchUsersSqlEnd(page, pageSize)).CollectPublicUserDTO(replicaConnection)
		if err != nil {
			return nil, err
		}
		return dtos, nil
	}

	entityFilters, err := UserFilterParser.ParseAll(rawFilters)
	if err != nil {
		return nil, err
	}

	filters, e := mapAndValidateFilters(entityFilters)
	if e != nil {
		return nil, Error.NewStatusError(e.Error(), http.StatusBadRequest)
	}

	valuesCount := 1
	// Not preallocated cuz if cond is condIsNull or condIsNotNull then there will be no value for this filter
	values := []any{}
	// query conds, not the filters ones
	conds := make([]string, len(filters))

	for i, filter := range filters {
		if filter.Property == user.PasswordProperty {
			return nil, Error.NewStatusError(
				"Can't search by user password",
				http.StatusBadRequest,
			)
		}

		conds[i] = filter.Build(valuesCount)

		if filter.Cond != condIsNotNull && filter.Cond != condIsNull {
			if filter.Value == nil || filter.Value == "" {
				return nil, Error.NewStatusError(
					"Filter has no value: " + rawFilters[i],
					http.StatusBadRequest,
				)
			}

			valuesCount++

			values = append(values, filter.Value)
		}
	}

    query := newQuery(searchUsersSqlStart + " WHERE " + strings.Join(conds, " AND ") + searchUsersSqlEnd(page, pageSize), values...)

    dtos, err := query.CollectPublicUserDTO(replicaConnection)
    if err != nil {
        return nil, err
    }

	return dtos, nil
}

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

	dbLogger.Trace("Searching for user with "+string(conditionProperty)+" = "+conditionValue+"...", nil)

	var query *query

	sql := `SELECT id, login, password, roles, deleted_at, version FROM "user" WHERE ` + string(conditionProperty) + ` = $1`

	switch state {
	case user.NotDeletedState:
		query = newQuery(sql + " AND deleted_at IS NULL;", conditionValue)
	case user.DeletedState:
		query = newQuery(sql + " AND deleted_at IS NOT NULL;", conditionValue)
	case user.AnyState:
		query = newQuery(sql + ";", conditionValue)
	default:
		dbLogger.Panic("Invalid findUserBy() call", "Unknown user state: " + string(state), nil)
		return nil, Error.StatusInternalError
	}

    dto, err := query.BasicUserDTO(replicaConnection, cacheKey)
    if err != nil {
        return nil, err
    }

	dbLogger.Trace("Searching for user with "+string(conditionProperty)+" = "+conditionValue+": OK", nil)

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

func (s *seeker) FindUserBySessionID(sessionID string) (*UserDTO.Basic, *Error.Status) {
	query := newQuery(
		`SELECT "user".id, "user".login, "user".password, "user".roles, "user".deleted_at, "user".version
		FROM "user" INNER JOIN "user_session" ON "user_session".user_id = "user".id
		WHERE "user_session".id = $1;`,
		sessionID,
	)

	cacheKey := cache.KeyBase[cache.UserBySessionID] + sessionID

	return query.BasicUserDTO(replicaConnection, cacheKey)
}

func (s *seeker) IsLoginAvailable(login string) bool  {
    if err := user.ValidateLogin(login); err != nil {
        return false
    }

    cacheKey := cache.KeyBase[cache.UserByLogin] + login

    _, err := s.findUserBy(user.LoginProperty, login, user.NotDeletedState, cacheKey)
    if err == nil {
        return false
    }

    return true
}

func (_ *seeker) GetRoles(act *ActionDTO.Targeted) ([]string, *Error.Status) {
    if err := act.ValidateTargetUID(); err != nil {
        return nil, err
    }

	if err := authz.User.GetUserRoles(act.RequesterRoles); err != nil {
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

    scan, err := query.Row(replicaConnection)

    if err != nil {
        return nil, err
    }

    roles := []string{}

    if e := scan(&roles); e != nil {
        return nil, e
    }

    cache.Client.Set(cache.KeyBase[cache.UserRolesById] + act.TargetUID, strings.Join(roles, ","))

    return roles, nil
}

func (_ *seeker) GetUserVersion(UID string) (uint32, *Error.Status) {
	cacheKey := cache.KeyBase[cache.UserVersionByID] + UID

	if cachedVersion, hit := cache.Client.Get(cacheKey); hit {
		ver, err := strconv.Atoi(cachedVersion)
		if err != nil {
			cache.Client.Delete(cacheKey)
		} else {
			return uint32(ver), nil
		}
	}

	query := newQuery(
		`SELECT version FROM "user" WHERE id = $1 AND deleted_at IS NULL;`,
		UID,
	)

	scan, err := query.Row(replicaConnection)
	if err != nil {
		return 0, err
	}

	var version uint32

	if err := scan(&version); err != nil {
		return 0, err
	}

	cache.Client.Set(cacheKey, version)

	return version, nil
}

