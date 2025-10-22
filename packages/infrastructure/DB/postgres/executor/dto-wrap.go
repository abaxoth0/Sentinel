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
	log "sentinel/packages/infrastructure/DB/postgres/logger"
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
	log.DB.Trace("Collecting rows...", nil)

	rows, err := Rows(conType, q)
	if err != nil {
		return nil, err
	}

	dtos, e := pgx.CollectRows(rows, collectFunc)
	if e != nil {
		log.DB.Error("Failed to collect rows", e.Error(), nil)
		return nil, q.ConvertAndLogError(e)
	}
	if len(dtos) == 0 {
		return nil, q.ConvertAndLogError(Error.StatusNotFound)
	}

	log.DB.Trace("Collecting rows: OK", nil)

	return dtos, nil
}

// TODO add cache
func CollectFullUserDTO(conType connection.Type, q *query.Query) ([]*UserDTO.Full, *Error.Status) {
	return collect(conType, q, func(row pgx.CollectableRow) (*UserDTO.Full, error) {
		dto := new(UserDTO.Full)

		var deletedAt sql.NullTime
		var createdAt sql.NullTime

		if err := row.Scan(
			&dto.ID,
			&dto.Login,
			&dto.Password,
			&dto.Roles,
			&deletedAt,
			&createdAt,
			&dto.Version,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			dto.DeletedAt = &deletedAt.Time
		}
		if createdAt.Valid {
			dto.CreatedAt = createdAt.Time
		}

		return dto, nil
	})
}

// TODO add cache
func CollectPublicUserDTO(conType connection.Type, q *query.Query) ([]*UserDTO.Public, *Error.Status) {
	return collect(conType, q, func(row pgx.CollectableRow) (*UserDTO.Public, error) {
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
// *UserDTO.Full after scanning resulting row into it.
func FullUserDTO(conType connection.Type, q *query.Query, cacheKey string) (*UserDTO.Full, *Error.Status) {
	if cached, hit := cache.Client.Get(cacheKey); hit {
		r, err := pbencoding.UnmarshallFullUserDTO([]byte(cached))
		if err == nil {
			return r, nil
		}

		// If decoding failed that means more likely cached data was invalid,
		// so need to delete it from cache to prevent errors in future.
		if e := cache.Client.Delete(cacheKey); e != nil {
			return nil, e
		}
	}

	scan, err := Row(conType, q)
	if err != nil {
		return nil, err
	}

	dto := new(UserDTO.Full)

	var deletedAt sql.NullTime
	var createdAt sql.NullTime

	if err := scan(
		&dto.ID,
		&dto.Login,
		&dto.Password,
		&dto.Roles,
		&deletedAt,
		&createdAt,
		&dto.Version,
	); err != nil {
		return nil, err
	}
	if deletedAt.Valid {
		dto.DeletedAt = &deletedAt.Time
	}
	if createdAt.Valid {
		dto.CreatedAt = createdAt.Time
	}

	if cached, e := pbencoding.MarshallFullUserDTO(dto); e == nil {
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

		// If decoding failed that means more likely cached data was invalid,
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

	if cached, e := pbencoding.MarshallFullSessionDTO(dto); e == nil {
		cache.Client.Set(cacheKey, cached)
	}

	return dto, nil
}

func CollectFullSessionDTO(conType connection.Type, query *query.Query) ([]*SessionDTO.Full, *Error.Status) {
	return collect(conType, query, func(row pgx.CollectableRow) (*SessionDTO.Full, error) {
		dto := new(SessionDTO.Full)

		var createdAt sql.NullTime
		var lastUsedAt sql.NullTime
		var expiresAt sql.NullTime
		var revokedAt sql.NullTime

		err := row.Scan(
			&dto.ID,
			&dto.UserID,
			&dto.UserAgent,
			&dto.IpAddress,
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

		return dto, nil
	})
}

func FullLocationDTO(conType connection.Type, q *query.Query, cacheKey string) (*LocationDTO.Full, *Error.Status) {
	if cached, hit := cache.Client.Get(cacheKey); hit {
		r, err := pbencoding.UnmarshallFullLocationDTO([]byte(cached))
		if err == nil {
			return r, nil
		}

		// If decoding failed that means more likely cached data was invalid,
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

	if cached, e := pbencoding.MarshallFullLocationDTO(dto); e == nil {
		cache.Client.Set(cacheKey, cached)
	}

	return dto, nil
}
