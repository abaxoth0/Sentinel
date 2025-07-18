package usertable

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/google/uuid"
)

func (m *Manager) Create(login string, password string) (string, *Error.Status) {
    if err := m.checkLogin(login); err != nil {
        return "", err
    }

    if err := user.ValidatePassword(password); err != nil {
        return "", err
    }

	hashedPassword, err := hashPassword(password)
    if err != nil {
        return "", nil
    }

    uid := uuid.New()

    insertQuery := query.New(
        `INSERT INTO "user" (id, login, password, roles) VALUES
        ($1, $2, $3, $4);`,
        uid, login, hashedPassword, rbac.GetRolesNames(authz.Host.DefaultRoles),
    )

    if err = cache.Client.DeleteOnNoError(
        executor.Exec(connection.Primary, insertQuery),
        cache.KeyBase[cache.UserByLogin] + login,
        cache.KeyBase[cache.AnyUserByLogin] + login,
    ); err != nil {
        return "", err
    }

    return uid.String(), nil
}

