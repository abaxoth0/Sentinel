package usertable

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func invalidateBasicUserDtoCache(old, current *UserDTO.Basic) {
	invalidator := cache.NewBasicUserDtoInvalidator(old, current)
	if err := invalidator.Invalidate(); err != nil {
		userLogger.Error("Failed to invalidate cache", err.Error(), nil)
	}
}

func (m *Manager) ChangeLogin(act *ActionDTO.UserTargeted, newLogin string) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

    if err := user.ValidateLogin(newLogin); err != nil {
        return err
    }

	if err := authz.User.ChangeUserLogin(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := m.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    if user.Login == newLogin {
        return Error.NewStatusError(
            "Login not changed: Current login and new login are the same",
            http.StatusConflict,
        )
    }

    if err := m.checkLogin(newLogin); err != nil {
        return err
    }

    audit := newAuditDTO(audit.UpdatedOperation, act, user)

    query := query.New(
        `UPDATE "user" SET login = $1, version = version + 1
        WHERE id = $2;`,
        newLogin, act.TargetUID,
    )

    if err := execTxWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.Login = newLogin
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

func (m *Manager) ChangePassword(act *ActionDTO.UserTargeted, newPassword string) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

    if err := user.ValidatePassword(newPassword); err != nil {
        return err
    }

	if err := authz.User.ChangeUserPassword(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := m.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

	hashedPassword, e := hashPassword(newPassword)
    if e != nil {
        return e
    }

    audit := newAuditDTO(audit.UpdatedOperation, act, user)

    query := query.New(
        `UPDATE "user" SET password = $1, version = version + 1
        WHERE id = $2;`,
        hashedPassword, act.TargetUID,
    )

    if err := execTxWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.Password = string(hashedPassword)
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

func (m *Manager) ChangeRoles(act *ActionDTO.UserTargeted, newRoles []string) *Error.Status {
    if err := act.ValidateUIDs(); err != nil {
        return err
    }

    if act.TargetUID == act.RequesterUID &&
       slices.Contains(act.RequesterRoles, "admin") &&
       !slices.Contains(newRoles, "admin") {
          return Error.NewStatusError(
              "Нельзя снять роль администратора с самого себя",
               http.StatusForbidden,
          )
    }

    user, err := m.FindUserByID(act.TargetUID)
    if err != nil {
        return err
    }

	if err := authz.User.ChangeUserRoles(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    audit := newAuditDTO(audit.UpdatedOperation, act, user)

    query := query.New(
        `UPDATE "user" SET roles = $1, version = version + 1
        WHERE id = $2;`,
        newRoles, act.TargetUID,
    )

    if err := execTxWithAudit(&audit, query); err != nil {
		return err
	}

	updatedUser := user.Copy()
	updatedUser.Roles = newRoles
	updatedUser.Version++
	invalidateBasicUserDtoCache(user, updatedUser)

	return nil
}

func (m *Manager) Activate(tk string) *Error.Status {
    t, err := token.ParseSingedToken(tk, config.Secret.ActivationTokenPublicKey)
    if err != nil {
        return err
    }

    payload, err := UserMapper.PayloadFromClaims(t.Claims.(jwt.MapClaims))
    if err != nil {
        return err
    }

    user, err := m.FindUserByID(payload.ID)
    if err != nil {
        return err
    }
    if user.IsActive() {
        return Error.NewStatusError(
            "User already active",
            http.StatusConflict,
        )
    }

    filter := ActionDTO.NewUserTargeted(user.ID, user.ID, user.Roles)

    audit := newAuditDTO(audit.UpdatedOperation, filter, user)
	var updatedUser *UserDTO.Basic

    for i, role := range user.Roles {
        if role == "unconfirmed_user" {
			updatedUser = user.Copy()
            updatedUser.Roles[i] = "user"
			updatedUser.Version++
            break
        }
    }

	// This should be impossible, but additional check won't be redundant
	if updatedUser == nil {
		userLogger.Error(
			"Failed to activate user " + user.ID,
			"User doesn't have role unconfirmed_user: " + strings.Join(user.Roles, ","),
			nil,
		)
		return Error.StatusInternalError
	}

    tx := transaction.New(
        query.New(
            `UPDATE "user" SET roles = $1, version = version + 1
             WHERE login = $2;`,
             updatedUser.Roles, updatedUser.Login,
        ),
        newAuditQuery(&audit),
    )
    if err := tx.Exec(connection.Primary); err != nil {
        return err
    }

	invalidateBasicUserDtoCache(user, updatedUser)

    return nil
}

