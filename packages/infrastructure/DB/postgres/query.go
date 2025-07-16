package postgres

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"reflect"
	"sentinel/packages/common/config"
	pbencoding "sentinel/packages/common/encoding/protobuf"
	Error "sentinel/packages/common/errors"
	LocationDTO "sentinel/packages/core/location/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/cache"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type query struct {
	prepared 	bool
    sql 		string
    args 		[]any
	// nil by default, can be initialized via query.prepare()
	con 		*pgxpool.Conn
	// nil by default, can be initialized via query.prepare()
	ctx 		context.Context
	// nil by default, can be initialized via query.prepare()
	cancel		context.CancelFunc
}

func newQuery(sql string, args ...any) *query {
    return &query{
		sql: sql,
		args: args,
	}
}

// Used to clear resources acquired for query (context, connection).
func (q *query) free() {
	q.con.Release()
	q.cancel()
}

func convertQueryError(err error, sql string) *Error.Status {
    defer dbLogger.Debug("Failed query: " + sql, nil)

    if err == context.DeadlineExceeded {
        dbLogger.Error("Query failed", "Query timeout", nil)
        return Error.StatusTimeout
    }

    dbLogger.Error("Query failed", err.Error(), nil)
    return Error.StatusInternalError
}

// Prepares query for execution by acquiring connection and creating contex.
// Also initializes q.free(), which is used to release connection and close context.
//
// Will cause panic if query was already prepared.
func(q *query) prepare(conType connectionType) (err *Error.Status) {
	if q.prepared {
		dbLogger.Panic("Failed to prepare query", "Query was already prepared", nil)
		return Error.StatusInternalError
	}

    con, err := driver.getConnection(conType)
    if err != nil {
        return err
    }

    ctx, cancel := defaultTimeoutContext()

	q.con = con
	q.ctx = ctx
	q.cancel = cancel

	if config.Debug.Enabled && config.Debug.LogDbQueries {
		args := make([]string, len(q.args))

		for i, arg := range q.args {
			switch a := arg.(type) {
			case string:
				args[i] = a
			case []string:
				args[i] = strings.Join(a, ", ")
			case int:
				args[i] = strconv.FormatInt(int64(a), 10)
			case int64:
				args[i] = strconv.FormatInt(a, 10)
			case int32:
				args[i] = strconv.FormatInt(int64(a), 10)
			case float32:
				args[i] = strconv.FormatFloat(float64(a), 'f', 8, 32)
			case float64:
				args[i] = strconv.FormatFloat(float64(a), 'f', 11, 64)
			case time.Time:
				args[i] = a.String()
			case bool:
				args[i] = strconv.FormatBool(a)
			case net.IP:
				args[i] = a.To4().String()
			}
		}

		dbLogger.Debug("Preparing query:\n" + q.sql + "\n * Query args: " + strings.Join(args, "; "), nil)
	}

	q.prepared = true

	return nil
}

func (q *query) Rows(conType connectionType) (pgx.Rows, *Error.Status) {
	if err := q.prepare(conType); err != nil {
		return nil, err
	}
	defer q.free()

	r, err := q.con.Query(q.ctx, q.sql, q.args...)
	if err != nil {
		return nil, convertQueryError(err, q.sql)
	}

	return r, nil
}

// Scans a row into the given destinations.
// All dests must be pointers.
// By default, dests validation is disabled,
// to enable this add "debug-safe-db-scans: true" to the config.
// (works only if app launched in debug mode)
type rowScanner = func(dests ...any) *Error.Status

// Wrapper for '*pgxpool.Con.QueryRow'
func (q *query) Row(conType connectionType) (rowScanner, *Error.Status) {
    if err := q.prepare(conType); err != nil{
        return nil, err
    }
	defer q.free()

	row := q.con.QueryRow(q.ctx, q.sql, q.args...)

    return func (dests ...any) *Error.Status {
		if config.Debug.Enabled && config.Debug.SafeDatabaseScans {
			for _, dest := range dests {
				typeof := reflect.TypeOf(dest)

				if typeof.Kind() != reflect.Ptr {
					dbLogger.Panic(
						"Query scan failed",
						"Destination for scanning must be a pointer, but got '"+typeof.String()+"'",
						nil,
					)
				}
			}
		}

		if e := row.Scan(dests...); e != nil {
			if errors.Is(e, pgx.ErrNoRows) {
				return Error.StatusNotFound
			}
			return convertQueryError(e, q.sql)
		}

		return nil
	}, nil
}

// Wrapper for '*pgxpool.Con.Exec'
func (q *query) Exec(conType connectionType) (*Error.Status) {
    if err := q.prepare(conType); err != nil {
		return err
	}
	defer q.free()

	if _, err := q.con.Exec(q.ctx, q.sql, q.args...); err != nil {
		return convertQueryError(err, q.sql)
	}

    return nil
}

// TODO add cache
func collect[T any](
	conType connectionType,
	q *query,
	collectFunc func(pgx.CollectableRow) (T, error),
) ([]T, *Error.Status) {
    rows, err := q.Rows(conType)
    if err != nil {
        return nil, err
    }

	dtos, e := pgx.CollectRows(rows, collectFunc)
    if e != nil {
		dbLogger.Error("Failed to collect rows", e.Error(), nil)
        return nil, convertQueryError(e, q.sql)
    }
	if len(dtos) == 0 {
		return nil, Error.StatusNotFound
	}

    return dtos, nil
}

// TODO add cache
func (q *query) CollectBasicUserDTO(conType connectionType) ([]*UserDTO.Basic, *Error.Status) {
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
func (q *query) CollectPublicUserDTO(conType connectionType) ([]*UserDTO.Public, *Error.Status) {
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
func (q *query) BasicUserDTO(conType connectionType, cacheKey string) (*UserDTO.Basic, *Error.Status) {
    if cached, hit := cache.Client.Get(cacheKey); hit {
		r, err := pbencoding.UnmarshallBasicUserDTO([]byte(cached))
        if err == nil {
            return r, nil
        }

        // if json decoding failed thats mean more likely it was invalid,
        // so deleting it from cache to prevent futher cache errors.
        // if it keep repeating even after this, then smth really went wrong.
        if e := cache.Client.Delete(cacheKey); e != nil {
            return nil, e
        }
    }

    scan, err := q.Row(conType)
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
		dbLogger.Error(
			"Failed to encode basic user DTO",
			e.Error(),
			nil,
		)
	} else {
    	cache.Client.Set(cacheKey, cached)
	}

    return dto, nil
}

// TODO add cache
func (q *query) FullSessionDTO(conType connectionType) (*SessionDTO.Full, *Error.Status) {
	scan, err := q.Row(conType)
	if err != nil {
		return nil, err
	}

	dto := new(SessionDTO.Full)

	var createdAt sql.NullTime
	var lastUsedAt sql.NullTime
	var expiresAt sql.NullTime
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
		&dto.Revoked,
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
}

func (q *query) CollectPublicSessionDTO(conType connectionType) ([]*SessionDTO.Public, *Error.Status) {
	return collect(conType, q, func (row pgx.CollectableRow) (*SessionDTO.Public, error) {
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

func (q *query) FullLocationDTO(conType connectionType) (*LocationDTO.Full, *Error.Status) {
	scan, err := q.Row(conType)
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

	return dto, nil
}

