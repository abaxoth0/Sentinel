package usertable

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserFilterParser "sentinel/packages/infrastructure/parsers/user-filter"
	"strconv"
	"strings"
)

const (
	searchUsersSqlStart=`WITH numbered_users AS (SELECT *, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) as row_num FROM "user"`
	searchUsersSqlSelect=`SELECT id, login, roles, deleted_at, version FROM numbered_users WHERE row_num BETWEEN `
)

func searchUsersSqlEnd(page, pageSize int) string {
	start := ((page - 1) * pageSize) + 1
	end := page * pageSize
	return ") " + searchUsersSqlSelect + strconv.Itoa(start) + " AND " + strconv.Itoa(end) + ";"
}

func (m *Manager) SearchUsers(
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
		searchQuery := query.New(searchUsersSqlStart + searchUsersSqlEnd(page, pageSize))
		dtos, err := executor.CollectPublicUserDTO(connection.Replica, searchQuery)
		if err != nil {
			return nil, err
		}
		return dtos, nil
	}

	entityFilters, err := UserFilterParser.ParseAll(rawFilters)
	if err != nil {
		return nil, err
	}

	filters, e := query.MapAndValidateUserFilters(entityFilters)
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

		if filter.Cond != query.CondIsNotNull && filter.Cond != query.CondIsNull {
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

    searchQuery := query.New(
		searchUsersSqlStart + " WHERE " + strings.Join(conds, " AND ") + searchUsersSqlEnd(page, pageSize), values...
	)

    dtos, err := executor.CollectPublicUserDTO(connection.Replica, searchQuery)
    if err != nil {
        return nil, err
    }

	return dtos, nil
}

func (m *Manager) checkLogin(login string) *Error.Status {
    if err := user.ValidateLogin(login); err != nil {
        return err
    }

    _, err := m.FindAnyUserByLogin(login)
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

func (m *Manager) findUserBy(
    conditionProperty user.Property,
    conditionValue string,
    state user.State,
    cacheKey string,
) (*UserDTO.Full, *Error.Status) {
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

	userLogger.Trace("Searching for user with "+string(conditionProperty)+" = "+conditionValue+"...", nil)

	var selectQuery *query.Query

	sql := `SELECT id, login, password, roles, deleted_at, created_at, version FROM "user" WHERE ` + string(conditionProperty) + ` = $1`

	switch state {
	case user.NotDeletedState:
		selectQuery = query.New(sql + " AND deleted_at IS NULL;", conditionValue)
	case user.DeletedState:
		selectQuery = query.New(sql + " AND deleted_at IS NOT NULL;", conditionValue)
	case user.AnyState:
		selectQuery = query.New(sql + ";", conditionValue)
	default:
		userLogger.Panic("Invalid findUserBy() call", "Unknown user state: " + string(state), nil)
		return nil, Error.StatusInternalError
	}

    dto, err := executor.FullUserDTO(connection.Replica, selectQuery, cacheKey)
    if err != nil {
        return nil, err
    }

	userLogger.Trace("Searching for user with "+string(conditionProperty)+" = "+conditionValue+": OK", nil)

    return dto, nil
}

func (m *Manager) FindAnyUserByID(id string) (*UserDTO.Full, *Error.Status) {
    return m.findUserBy(
        user.IdProperty,
        id,
        user.AnyState,
        cache.KeyBase[cache.AnyUserById] + id,
    )
}

func (m *Manager) FindUserByID(id string) (*UserDTO.Full, *Error.Status) {
    return m.findUserBy(
        user.IdProperty,
        id,
        user.NotDeletedState,
        cache.KeyBase[cache.UserById] + id,
    )
}

func (m *Manager) FindSoftDeletedUserByID(id string) (*UserDTO.Full, *Error.Status) {
    return m.findUserBy(
        user.IdProperty,
        id,
        user.DeletedState,
        cache.KeyBase[cache.DeletedUserById] + id,
    )
}

func (m *Manager) FindAnyUserByLogin(login string) (*UserDTO.Full, *Error.Status) {
    return m.findUserBy(
        user.LoginProperty,
        login,
        user.AnyState,
        cache.KeyBase[cache.AnyUserByLogin] + login,
    )
}

func (m *Manager) FindUserByLogin(login string) (*UserDTO.Full, *Error.Status) {
    return m.findUserBy(
        user.LoginProperty,
        login,
        user.NotDeletedState,
        cache.KeyBase[cache.UserByLogin] + login,
    )
}

func (m *Manager) FindUserBySessionID(sessionID string) (*UserDTO.Full, *Error.Status) {
	selectQuery := query.New(
		`SELECT "user".id, "user".login, "user".password, "user".roles, "user".deleted_at, "user".created_at, "user".version
		FROM "user" INNER JOIN "user_session" ON "user_session".user_id = "user".id
		WHERE "user_session".id = $1;`,
		sessionID,
	)

	cacheKey := cache.KeyBase[cache.UserBySessionID] + sessionID

	dto, err := executor.FullUserDTO(connection.Replica, selectQuery, cacheKey)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (m *Manager) IsLoginAvailable(login string) bool  {
    if err := user.ValidateLogin(login); err != nil {
        return false
    }

    cacheKey := cache.KeyBase[cache.UserByLogin] + login

    _, err := m.findUserBy(user.LoginProperty, login, user.NotDeletedState, cacheKey)
    if err == nil {
        return false
    }

    return true
}

func (_ *Manager) GetRoles(act *ActionDTO.UserTargeted) ([]string, *Error.Status) {
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
    selectQuery := query.New(
        `SELECT roles FROM "user" WHERE id = $1 AND deleted_at IS NULL;`,
        act.TargetUID,
    )

    scan, err := executor.Row(connection.Replica, selectQuery)

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

func (_ *Manager) GetUserVersion(UID string) (uint32, *Error.Status) {
	cacheKey := cache.KeyBase[cache.UserVersionByID] + UID

	if cachedVersion, hit := cache.Client.Get(cacheKey); hit {
		ver, err := strconv.Atoi(cachedVersion)
		if err != nil {
			cache.Client.Delete(cacheKey)
		} else {
			return uint32(ver), nil
		}
	}

	selectQuery := query.New(
		`SELECT version FROM "user" WHERE id = $1 AND deleted_at IS NULL;`,
		UID,
	)

	scan, err := executor.Row(connection.Replica, selectQuery)
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

