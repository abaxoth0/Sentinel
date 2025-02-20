package postgres

import (
	"errors"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errs"

	"github.com/jackc/pgx/v5"
)

// Executes given sql.
// Substitutes given args (see pgx docs for details).
func evalSQL(sql string, args ...any) (pgx.Row, *Error.Status) {
    con, err := driver.getConnection()

    defer con.Release()

    if err != nil {
        return nil, err
    }

    return con.QueryRow(
        driver.ctx,
        sql,
        args...,
    ), nil

}

func dtoFromQuery(sql string, args ...any) (*UserDTO.Indexed, *Error.Status) {
    dto := new(UserDTO.Indexed)

    row, err := evalSQL(sql, args)

    if err != nil {
        return nil, err
    }

    e := row.Scan(&dto.ID, &dto.Login, &dto.Password, &dto.Roles, &dto.DeletedAt)

    if e != nil {
        if errors.Is(e, pgx.ErrNoRows) {
            return nil, Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
        }

        println(e.Error())

        return nil, Error.NewStatusError("Internal Server Error", http.StatusInternalServerError)
    }

    return dto, nil
}


