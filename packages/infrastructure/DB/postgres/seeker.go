package postgres

import (
	"context"
	"errors"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type seeker struct {
    //
}

var invalidUID = Error.NewStatusError("Invalid ID", http.StatusBadRequest)

func (_ *seeker) FindUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    parsedID, e := strconv.ParseInt(id, 10, 64);

    if e != nil {
        return nil, invalidUID
    }

    return queryDTO(
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE id = $1 AND deletedAt IS NULL;`,
        parsedID,
    )
}

func (_ *seeker) FindSoftDeletedUserByID(id string) (*UserDTO.Indexed, *Error.Status) {
    parsedID, e := strconv.ParseInt(id, 10, 64);

    if e != nil {
        return nil, invalidUID
    }

    return queryDTO(
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE id = $1 AND deletedAt IS NOT NULL;`,
        parsedID,
    )
}

func (_ *seeker) FindUserByLogin(login string) (*UserDTO.Indexed, *Error.Status) {
    return queryDTO(
        `SELECT id, login, password, roles, deletedAt
         FROM "user"
         WHERE login = $1 AND deletedAt IS NULL;`,
        login,
    )
}

func (_ *seeker) IsLoginExists(target string) (bool, *Error.Status) {
    var id int

    row, err := queryRow(
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
        if err == context.DeadlineExceeded {
            return false, Error.StatusTimeout
        }

        if errors.Is(e, pgx.ErrNoRows) {
            return false, nil
        }

        println(e.Error())

        return false, Error.StatusInternalError
    }

    return true, nil
}

