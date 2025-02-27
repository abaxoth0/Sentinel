package postgres

import (
	"context"
	"errors"
	"fmt"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"

	"github.com/jackc/pgx/v5"
)

// Executes given SQL. If returnRow is true then returns resulting row and error,
// otherwise returns nil and error.
// Also substitutes given args (see pgx docs for details).
func evalSQL(returnRow bool, sql string, args []any) (pgx.Row, *Error.Status) {
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
        if e == context.DeadlineExceeded {
            return nil, Error.StatusTimeout
        }

        fmt.Printf("[ ERROR ] Failed to execute query `%s`: \n%v\n", sql, e.Error())

        return nil, Error.StatusInternalError
    }

    return nil, nil
}

// Wrapper for '*pgxpool.Con.QueryRow'
func queryRow(sql string, args ...any) (pgx.Row, *Error.Status) {
    return evalSQL(true, sql, args)
}

// Wrapper for '*pgxpool.Con.Exec'
func queryExec(sql string, args ...any) (*Error.Status) {
    _, err := evalSQL(false, sql, args)

    return err
}

// Works same as 'queryRow', but also scans resulting row into '*UserDTO.Indexed'
func queryDTO(sql string, args ...any) (*UserDTO.Indexed, *Error.Status) {
    dto := new(UserDTO.Indexed)

    row, err := queryRow(sql, args)

    if err != nil {
        return nil, err
    }

    e := row.Scan(&dto.ID, &dto.Login, &dto.Password, &dto.Roles, &dto.DeletedAt)

    if e != nil {
        if e == context.DeadlineExceeded {
            return nil, Error.StatusTimeout
        }

        if errors.Is(e, pgx.ErrNoRows) {
            return nil, Error.StatusUserNotFound
        }

        fmt.Printf("[ ERROR ] Failed to execute query `%s`: \n%v\n", sql, e.Error())

        return nil, Error.StatusInternalError
    }

    return dto, nil
}

