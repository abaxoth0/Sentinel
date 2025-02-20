package postgres

import (
	"errors"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errs"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type seeker struct {
    //
}

func (_ *seeker) FindUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    parsedID, e := strconv.ParseInt(id, 10, 64);

    if e != nil {
        return nil, Error.NewStatusError("Invalid ID", http.StatusBadRequest)
    }

    return dtoFromQuery(
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE id = $1 AND deletedAt IS NULL;`,
        parsedID,
    )
}

func (_ *seeker) FindSoftDeletedUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    parsedID, e := strconv.ParseInt(id, 10, 64);

    if e != nil {
        return nil, Error.NewStatusError("Invalid ID", http.StatusBadRequest)
    }

    return dtoFromQuery(
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE id = $1 AND deletedAt IS NOT NULL;`,
        parsedID,
    )
}

func (_ *seeker) FindUserByLogin(login string) (*UserDTO.Indexed, *Error.Status) {
    return dtoFromQuery(
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE login = $1 AND deletedAt IS NULL;`,
        login,
    )
}

func (_ *seeker) IsLoginExists(target string) (bool, *Error.Status) {
    var id int
    row, err := evalSQL(
        `SELECT id
         FROM "user"
         WHERE login = $1 and deletedAt IS NULL;`,
        target,
    )

    if err != nil {
        return false, err
    }

    e := row.Scan(&id)

    if e != nil {
        if errors.Is(e, pgx.ErrNoRows) {
            return false, nil
        }

        println(e.Error())

        return false, Error.NewStatusError("Internal Server Error", http.StatusInternalServerError)
    }

    return true, nil
}

