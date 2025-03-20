package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/presentation/data/json"
	"time"

	"github.com/jackc/pgx/v5"
)

func logQueryError(query string, err error) *Error.Status {
    if err == context.DeadlineExceeded {
        fmt.Printf("[ ERROR ] Query timeout:\n%s\n", query)
        return Error.StatusTimeout
    }

    fmt.Printf("[ ERROR ] Failed to execute query `%s`: \n%v\n", query, err.Error())

    return Error.StatusInternalError
}

// Executes given SQL. If returnRow is true then returns resulting row and error,
// otherwise returns nil and error.
// Also substitutes given args (see pgx docs for details).
func runSQL(returnRow bool, query string, args []any) (pgx.Row, *Error.Status) {
    con, err := driver.getConnection()

    if err != nil {
        return nil, err
    }

    defer con.Release()

    ctx, cancel := defaultTimeoutContext()

    defer cancel()

    if returnRow {
        return con.QueryRow(ctx, query, args...), nil
    }

    if _, e := con.Exec(ctx, query, args...); e != nil {
        return nil, logQueryError(query, e)
    }

    return nil, nil
}

// Scans a row into the given destinations.
// All args should be a pointers.
// If safe is true, then all dests will be validated to be a pointers
type scanRow = func(safe bool, dests ...any) *Error.Status

// Wrapper for '*pgxpool.Con.QueryRow'
func queryRow(query string, args ...any) (scanRow, *Error.Status) {
    row, err := runSQL(true, query, args)

    if err != nil {
        return nil, err
    }

    return func (safe bool, dests ...any) *Error.Status {
        if safe {
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

            return logQueryError(query, e)
        }

        return nil
    }, nil
}

// Wrapper for '*pgxpool.Con.Exec'
func queryExec(sql string, args ...any) (*Error.Status) {
    _, err := runSQL(false, sql, args)

    return err
}

// Works same as 'queryRow', but also scans resulting row into '*UserDTO.Indexed'
func queryBasicDTO(cacheKey string, query string, args ...any) (*UserDTO.Basic, *Error.Status) {
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

    scan, err := queryRow(query, args...)

    if err != nil {
        return nil, err
    }

    dto := new(UserDTO.Basic)

    var deletedAt sql.NullTime

    err = scan(false, &dto.ID, &dto.Login, &dto.Password, &dto.Roles, &deletedAt)

    if err != nil {
        return nil, err
    }

    if deletedAt.Valid {
        dto.DeletedAt = deletedAt.Time
    } else {
        dto.DeletedAt = time.Time{}
    }


    cache.Client.EncodeAndSet(cacheKey, dto)

    return dto, nil
}

