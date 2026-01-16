package usertable

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/dblog"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"

	rbac "github.com/abaxoth0/SentinelRBAC"
	"github.com/google/uuid"
)

func (m *Manager) Create(login string, password string) (string, *Error.Status) {
	dblog.Logger.Info("Creating new user...", nil)

	if err := m.checkLoginAvailability(login); err != nil {
		return "", err
	}

	if err := user.ValidatePassword(password); err != nil {
		dblog.Logger.Error("Failed to create new user", err.Error(), nil)
		return "", err
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		dblog.Logger.Error("Failed to create new user", err.Error(), nil)
		return "", nil
	}

	uid := uuid.New()

	insertQuery := query.New(
		`INSERT INTO "user" (id, login, password, roles) VALUES
        ($1, $2, $3, $4);`,
		uid, login, hashedPassword, rbac.GetRolesNames(authz.Host.DefaultRoles),
	)

	if err := executor.Exec(connection.Primary, insertQuery); err != nil {
		return "", err
	}

	cache.Client.Delete(
		cache.KeyBase[cache.UserByLogin]+login,
		cache.KeyBase[cache.AnyUserByLogin]+login,
	)

	dblog.Logger.Info("Creating new user: OK", nil)

	return uid.String(), nil
}
