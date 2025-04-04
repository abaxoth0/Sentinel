package postgres

import (
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/core/activation"
	UserDTO "sentinel/packages/core/user/DTO"
	"time"

	"github.com/google/uuid"
)

func newActivationQuery(login string) *query {
    return newQuery(
        `INSERT INTO "user_activation" (user_login, token, expires_at)
         VALUES ($1, $2, $3);`,
        login, uuid.New(), time.Now().Add(config.App.ActivationTokenTTL()),
    )
}

func newDeleteActivationQuery(property activation.Property, value string) *query {
    return newQuery(
        `DELETE FROM "user_activation"
         WHERE `+ string(property) + ` = $1;`,
        value,
    )
}

func (_ *repository) Activate(token string) *Error.Status {
    activ, err := driver.FindActivationByToken(token)

    if err != nil {
        return err
    }

    if activ.ExpiresAt.Compare(time.Now()) == -1 {
        return activationTokenExpired
    }

    user, err := driver.FindUserByLogin(activ.UserLogin)

    if err != nil {
        return err
    }

    filter :=  &UserDTO.Filter{
        TargetUID: user.ID,
        RequesterUID: user.ID,
        RequesterRoles: user.Roles,
    }

    audit := newAudit(updatedOperation, filter, user)

    tx := newTransaction(
        newQuery(
            `UPDATE "user" SET is_active = true
             WHERE login = $1 AND is_active = false;`,
             activ.UserLogin,
        ),
        newDeleteActivationQuery(activation.TokenProperty, token),
        newAuditQuery(&audit),
    )

    if err := tx.Exec(); err != nil {
        return err
    }

    return nil
}

func (_ *repository) Reactivate(login string) *Error.Status {
    _, err := driver.FindUserByLogin(login)

    if err != nil {
        return err
    }

    activ, err := driver.FindActivationByUserLogin(login)

    if err != nil {
        return err
    }

    tx := newTransaction(
        newDeleteActivationQuery(activation.TokenProperty, activ.Token),
        newActivationQuery(login),
    )

    if err := tx.Exec(); err != nil {
        return err
    }

    return nil
}

