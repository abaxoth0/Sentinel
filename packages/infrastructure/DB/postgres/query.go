package postgres

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"

	"github.com/jackc/pgx/v5"
)

func handleQueryError(query string, err error) *Error.Status {
    if err == context.DeadlineExceeded {
        return Error.StatusTimeout
    }

    fmt.Printf("[ ERROR ] Failed to execute query `%s`: \n%v\n", query, err.Error())

    return Error.StatusInternalError
}

// Executes given SQL. If returnRow is true then returns resulting row and error,
// otherwise returns nil and error.
// Also substitutes given args (see pgx docs for details).
func runSQL(returnRow bool, sql string, args []any) (pgx.Row, *Error.Status) {
    con, err := driver.getConnection()

    if err != nil {
        return nil, err
    }

    defer con.Release()

    ctx, cancel := defaultTimeoutContext()

    defer cancel()

    if returnRow {
        return con.QueryRow(ctx, sql, args...), nil
    }

    if _, e := con.Exec(ctx, sql, args...); e != nil {
        return nil, handleQueryError(sql, e)
    }

    return nil, nil
}

// Scans a row into the given destinations.
// All args should be a pointers.
// If safe is true, then all dests will be validated to be a pointers
type scanRow = func(safe bool, dests ...any) *Error.Status

// Wrapper for '*pgxpool.Con.QueryRow'
func queryRow(sql string, args ...any) (scanRow, *Error.Status) {
    row, err := runSQL(true, sql, args)

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

            return handleQueryError(sql, e)
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
func queryDTO(sql string, args ...any) (*UserDTO.Indexed, *Error.Status) {
    scan, err := queryRow(sql, args...)

    if err != nil {
        return nil, err
    }

    dto := new(UserDTO.Indexed)

    return dto, scan(true, &dto.ID, &dto.Login, &dto.Password, &dto.Roles, &dto.DeletedAt)
}

