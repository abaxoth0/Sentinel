package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/common/config"
	"sentinel/packages/presentation/data/json"

	"github.com/jackc/pgx/v5"
)

type query struct {
    sql string
    args []any
}

func newQuery(sql string, args ...any) *query {
    return &query{sql, args}
}

func (q *query) toStatusError(err error) *Error.Status {
    if err == context.DeadlineExceeded {
        fmt.Printf("[ ERROR ] Query timeout:\n%s\n", q.sql)
        return Error.StatusTimeout
    }

    fmt.Printf("[ ERROR ] Failed to execute query `%s`: \n%v\n", q.sql, err.Error())
    return Error.StatusInternalError
}

// Executes given SQL. If returnRow is true then returns resulting row and error,
// otherwise returns nil and error.
// Also substitutes query args (see pgx docs for details).
func(q *query) runSQL(returnRow bool) (pgx.Row, *Error.Status) {
    con, err := driver.getConnection()

    if err != nil {
        return nil, err
    }

    defer con.Release()

    ctx, cancel := defaultTimeoutContext()

    defer cancel()

    if returnRow {
        return con.QueryRow(ctx, q.sql, q.args...), nil
    }

    if _, e := con.Exec(ctx, q.sql, q.args...); e != nil {
        return nil, q.toStatusError(e)
    }

    return nil, nil
}

// Scans a row into the given destinations.
// All dests must be pointers.
// By default, dests are not validated,
// but it can be added by setting env variable DEBUG_SAFE_DB_SCANS to true.
// (works only if app launched in debug mode)
type scanRow = func(dests ...any) *Error.Status

// Wrapper for '*pgxpool.Con.QueryRow'
func (q *query) Row() (scanRow, *Error.Status) {
    row, err := q.runSQL(true)

    if err != nil {
        return nil, err
    }

    return func (dests ...any) *Error.Status {
        if config.Debug.Enabled && config.Debug.SafeDatabaseScans {
            for _, dest := range dests {
                typeof := reflect.TypeOf(dest)

                if typeof.Kind() != reflect.Ptr {
                    fmt.Printf("[ ERROR ] Destination for scanning must be a pointer, got: %s\n", typeof.String())
                    return Error.StatusInternalError
                }
            }
        }

        if e := row.Scan(dests...); e != nil {
            if errors.Is(e, pgx.ErrNoRows) {
                // IMPORTANT since DB contains only users we
                //           can consider that exactly user wans't found,
                //           but if some other entity will be added to DB
                //           this error must be change onto smth like
                //           Error.StatusNotFound or Error.StatusNoResult
                return Error.StatusUserNotFound
            }

            return q.toStatusError(e)
        }

        return nil
    }, nil
}

// Wrapper for '*pgxpool.Con.Exec'
func (q *query) Exec() (*Error.Status) {
    _, err := q.runSQL(false)

    return err
}

// Works same as queryRow, but also creates and returns
// UserDTO.Basic after scanning resulting row into it.
func (q *query) RowBasicUserDTO(cacheKey string) (*UserDTO.Basic, *Error.Status) {
    if cached, hit := cache.Client.Get(cacheKey); hit {
        r, err := json.DecodeString[UserDTO.Basic](cached)

        if err == nil {
            return &r, nil
        }

        // if json decoding failed thats mean more likely it was invalid,
        // so deleting it from cache to prevent futher cache errors.
        // if it keep repeating even after this, then smth really went wrong in json package.
        cache.Client.Delete(cacheKey)
    }

    scan, err := q.Row()

    if err != nil {
        return nil, err
    }

    dto := new(UserDTO.Basic)

    var deletedAt sql.NullTime

    err = scan(
        &dto.ID,
        &dto.Login,
        &dto.Password,
        &dto.Roles,
        &deletedAt,
        &dto.IsActive,
    )

    setTime(&dto.DeletedAt, deletedAt)

    if err != nil {
        return nil, err
    }

    cache.Client.EncodeAndSet(cacheKey, dto)

    return dto, nil
}

