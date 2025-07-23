package executor

import (
	"database/sql"
	"net"
	pbencoding "sentinel/packages/common/encoding/protobuf"
	Error "sentinel/packages/common/errors"
	LocationDTO "sentinel/packages/core/location/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/cache"

	"github.com/jackc/pgx/v5"
)

// TODO a lot of code duplication

// TODO add cache
func collect[T any](
	conType connection.Type,
	q *query.Query,
	collectFunc func(pgx.CollectableRow) (T, error),
) ([]T, *Error.Status) {
    rows, err := Rows(conType, q)
    if err != nil {
        return nil, err
    }

	dtos, e := pgx.CollectRows(rows, collectFunc)
    if e != nil {
		executorLogger.Error("Failed to collect rows", e.Error(), nil)
        return nil, q.ConvertError(e)
    }
	if len(dtos) == 0 {
		return nil, Error.StatusNotFound
	}

    return dtos, nil
}

// TODO add cache
func CollectBasicUserDTO(conType connection.Type, q *query.Query) ([]*UserDTO.Basic, *Error.Status) {
	return collect(conType, q, func (row pgx.CollectableRow) (*UserDTO.Basic, error) {
		dto := new(UserDTO.Basic)

		var deletedAt sql.NullTime

		if err := row.Scan(
			&dto.ID,
			&dto.Login,
			&dto.Password,
			&dto.Roles,
			&deletedAt,
			&dto.Version,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			dto.DeletedAt = deletedAt.Time
		}

		return dto, nil
	})
}

// TODO add cache
func CollectPublicUserDTO(conType connection.Type, q *query.Query) ([]*UserDTO.Public, *Error.Status) {
	return collect(conType, q, func (row pgx.CollectableRow) (*UserDTO.Public, error) {
		dto := new(UserDTO.Public)

		var deletedAt sql.NullTime

		if err := row.Scan(
			&dto.ID,
			&dto.Login,
			&dto.Roles,
			&deletedAt,
			&dto.Version,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			dto.DeletedAt = &deletedAt.Time
		}

		return dto, nil
	})
}

// Works same as queryRow, but also creates and returns
// UserDTO.Basic after scanning resulting row into it.
func BasicUserDTO(conType connection.Type, q *query.Query, cacheKey string) (*UserDTO.Basic, *Error.Status) {
    if cached, hit := cache.Client.Get(cacheKey); hit {
		r, err := pbencoding.UnmarshallBasicUserDTO([]byte(cached))
        if err == nil {
            return r, nil
        }

        // If decoding failed thats mean more likely cached data was invalid,
        // so need to delete it from cache to prevent errors in future.
        if e := cache.Client.Delete(cacheKey); e != nil {
            return nil, e
        }
    }

    scan, err := Row(conType, q)
    if err != nil {
        return nil, err
    }

	dto := new(UserDTO.Basic)

	var deletedAt sql.NullTime

	if err := scan(
		&dto.ID,
		&dto.Login,
		&dto.Password,
		&dto.Roles,
		&deletedAt,
		&dto.Version,
	); err != nil {
		return nil, err
	}
	if deletedAt.Valid {
		dto.DeletedAt = deletedAt.Time
	}

	cached, e := pbencoding.MarshallBasicUserDTO(dto)
	if e != nil {
		executorLogger.Error("Failed to encode basic user DTO", e.Error(), nil)
	} else {
    	cache.Client.Set(cacheKey, cached)
	}

    return dto, nil
}

func FullSessionDTO(conType connection.Type, q *query.Query, cacheKey string) (*SessionDTO.Full, *Error.Status) {
	if cached, hit := cache.Client.Get(cacheKey); hit {
		r, err := pbencoding.UnmarshallFullSessionDTO([]byte(cached))
		if err == nil {
			return r, nil
		}

        // If decoding failed thats mean more likely cached data was invalid,
        // so need to delete it from cache to prevent same errors in future.
        if e := cache.Client.Delete(cacheKey); e != nil {
            return nil, e
        }
	}

	scan, err := Row(conType, q)
	if err != nil {
		return nil, err
	}

	dto := new(SessionDTO.Full)

	var createdAt sql.NullTime
	var lastUsedAt sql.NullTime
	var expiresAt sql.NullTime
	var revokedAt sql.NullTime
	var addr net.IP

	err = scan(
		&dto.ID,
		&dto.UserID,
		&dto.UserAgent,
		&addr,
		&dto.DeviceID,
		&dto.DeviceType,
		&dto.OS,
		&dto.OSVersion,
		&dto.Browser,
		&dto.BrowserVersion,
		&createdAt,
		&lastUsedAt,
		&expiresAt,
		&revokedAt,
	)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		dto.CreatedAt = createdAt.Time
	}
	if lastUsedAt.Valid {
		dto.LastUsedAt = lastUsedAt.Time
	}
	if expiresAt.Valid {
		dto.ExpiresAt = expiresAt.Time
	}
	if revokedAt.Valid {
		dto.RevokedAt = revokedAt.Time
	}
	dto.IpAddress = addr

	cached, e := pbencoding.MarshallFullSessionDTO(dto)
	if e != nil {
		executorLogger.Error("Failed to encode full session DTO", e.Error(), nil)
	} else {
    	cache.Client.Set(cacheKey, cached)
	}

	return dto, nil
}

func CollectPublicSessionDTO(conType connection.Type, query *query.Query) ([]*SessionDTO.Public, *Error.Status) {
	return collect(conType, query, func (row pgx.CollectableRow) (*SessionDTO.Public, error) {
		dto := new(SessionDTO.Public)

		var createdAt sql.NullTime
		var lastUsedAt sql.NullTime
		var expiresAt sql.NullTime
		var addr net.IP

		err := row.Scan(
			&dto.ID,
			&dto.UserAgent,
			&addr,
			&dto.DeviceID,
			&dto.DeviceType,
			&dto.OS,
			&dto.OSVersion,
			&dto.Browser,
			&dto.BrowserVersion,
			&createdAt,
			&lastUsedAt,
			&expiresAt,
		)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			dto.CreatedAt = createdAt.Time
		}
		if lastUsedAt.Valid {
			dto.LastUsedAt = lastUsedAt.Time
		}
		if expiresAt.Valid {
			dto.ExpiresAt = expiresAt.Time
		}
		dto.IpAddress = addr.To4().String()

		return dto, nil
	})
}

func FullLocationDTO(conType connection.Type, q *query.Query, cacheKey string) (*LocationDTO.Full, *Error.Status) {
	if cached, hit := cache.Client.Get(cacheKey); hit {
		r, err := pbencoding.UnmarshallFullLocationDTO([]byte(cached))
		if err == nil {
			return r, nil
		}

        // If decoding failed thats mean more likely cached data was invalid,
        // so need to delete it from cache to prevent same errors in future.
        if e := cache.Client.Delete(cacheKey); e != nil {
            return nil, e
        }
	}

	scan, err := Row(conType, q)
	if err != nil {
		return nil, err
	}

	dto := new(LocationDTO.Full)

	var createdAt sql.NullTime
	var deletedAt sql.NullTime
	var addr net.IP

	err = scan(
		&dto.ID,
		&addr,
		&dto.SessionID,
		&dto.Country,
		&dto.Region,
		&dto.City,
		&dto.Latitude,
		&dto.Longitude,
		&dto.ISP,
		&deletedAt,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		dto.CreatedAt = createdAt.Time
	}
	if deletedAt.Valid {
		dto.DeletedAt = deletedAt.Time
	}
	dto.IP = addr

	cached, e := pbencoding.MarshallFullLocationDTO(dto)
	if e != nil {
		executorLogger.Error("Failed to encode full location DTO", e.Error(), nil)
	} else {
    	cache.Client.Set(cacheKey, cached)
	}

	return dto, nil
}

