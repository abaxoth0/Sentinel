package postgres

import (
	"fmt"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/filter"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	"strings"
)

type seeker struct {
    //
}

func (s *seeker) SearchUsers(
	act *ActionDTO.Basic,
	entityFilters []filter.Entity[user.Property],
) ([]*UserDTO.Public, *Error.Status) {
	if err := authz.User.SearchUsers(act.RequesterRoles); err != nil {
		return nil, err
	}

	if entityFilters == nil || len(entityFilters) == 0 {
		dbLogger.Panic(
			"Failed to find users",
			fmt.Sprintf("Invalid filters value, expected non-nil and non-empty slice, but got: %+v", entityFilters),
			nil,
		)
		return nil, Error.StatusInternalError
	}

	filters := mapFilters(entityFilters)

	sql := `SELECT id, login, roles, deleted_at FROM "user" WHERE `

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
		if filter.Property == user.IdProperty {
			if err := validation.UUID(filter.StringValue()); err != nil {
				return nil, err.ToStatus(
					"user id isn't specified",
					"user id has invalid value",
				)
			}
		}
		if filter.Property == user.LoginProperty && config.App.IsLoginEmail {
			if err := validation.Email(filter.StringValue()); err != nil {
				return nil, err.ToStatus(
					"User login isn't specified",
					"User login has invalid value",
				)
			}
		}

		conds[i] = filter.Build(valuesCount)

		if filter.Cond != condIsNotNull && filter.Cond != condIsNull {
			if filter.Value == nil {
				dbLogger.Panic(
					"Failed to find user",
					"Filter value is nil",
					nil,
				)
				return nil, Error.StatusInternalError
			}

			valuesCount++

			values = append(values, filter.Value)
		}
	}

    query := newQuery(sql + strings.Join(conds, " AND ") + ";", values...)

    dtos, err := query.CollectPublicUserDTO()
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

    query := newQuery(
        `SELECT id, login, password, roles, deleted_at
        FROM "user"
        WHERE ` + string(conditionProperty) + ` = $1;`,
        conditionValue,
    )

    dto, err := query.BasicUserDTO(cacheKey)

    if err != nil {
        return nil, err
    }

    if state == user.NotDeletedState && dto.IsDeleted() {
        return nil, Error.StatusNotFound
    }

    if state == user.DeletedState && !dto.IsDeleted() {
        return nil, Error.StatusNotFound
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

    scan, err := query.Row()

    if err != nil {
        return nil, err
    }

    roles := []string{}

    if e := scan(&roles); e != nil {
        if e == Error.StatusNotFound {
            return nil, Error.StatusNotFound
        }
        return nil, e
    }

    cache.Client.Set(cache.KeyBase[cache.UserRolesById] + act.TargetUID, strings.Join(roles, ","))

    return roles, nil
}

