package postgres

import (
	"database/sql"
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"sentinel/packages/core/activation"
	ActivationDTO "sentinel/packages/core/activation/DTO"
)

// TODO Add cache
// TODO Make value work with different types
func (_ *seeker) findActivationBy(property activation.Property, value string) (*ActivationDTO.Basic, *Error.Status) {
    query := newQuery(
        `SELECT id, user_login, token, expires_at, created_at FROM "user_activation"
         WHERE `+ string(property) + ` = $1`,
         value,
    )

    scan, err := query.Row()

    if err != nil {
        return nil, err
    }

    activation := new(ActivationDTO.Basic)

    var expiresAt sql.NullTime
    var createdAt sql.NullTime

    if err := scan(
        &activation.Id,
        &activation.UserLogin,
        &activation.Token,
        &expiresAt,
        &createdAt,
    ); err != nil {
        if err == Error.StatusNotFound {
            return nil, activationNotFound
        }
        return nil, err
    }

    setTime(&activation.ExpiresAt, expiresAt)
    setTime(&activation.CreatedAt, createdAt)

    return activation, nil
}

func (_ *seeker) FindActivationByToken(token string) (*ActivationDTO.Basic, *Error.Status) {
    if err := validation.UUID(token); err != nil {
        if err == Error.NoValue {
            return nil, Error.NewStatusError("Activation token is not specified", http.StatusBadRequest)
        }
        if err == Error.InvalidValue {
            return nil, invalidActivationTokenFormat
        }
    }

    return driver.findActivationBy(activation.TokenProperty, token)
}

func (_ *seeker) FindActivationByUserLogin(userLogin string) (*ActivationDTO.Basic, *Error.Status) {
    return driver.findActivationBy(activation.UserLoginProperty, userLogin)
}

