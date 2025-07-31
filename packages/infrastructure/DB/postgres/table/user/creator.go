package usertable

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/google/uuid"
)

func (m *Manager) Create(login string, password string) (string, *Error.Status) {
	log.DB.Info("Creating new user...", nil)

    if err := m.checkIfLoginInUse(login); err != nil {
        return "", err
    }

    if err := user.ValidatePassword(password); err != nil {
		log.DB.Error("Failed to create new user", err.Error(), nil)
        return "", err
    }

	hashedPassword, err := hashPassword(password)
    if err != nil {
		log.DB.Error("Failed to create new user", err.Error(), nil)
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

	// TODO Handle error (need to do that for same cases).
	// 		Create queue (or two) which will try to clear cache for this kinda "dirty" keys?
    cache.Client.Delete(
        cache.KeyBase[cache.UserByLogin] + login,
        cache.KeyBase[cache.AnyUserByLogin] + login,
    )

	log.DB.Info("Creating new user: OK", nil)

    return uid.String(), nil
}

