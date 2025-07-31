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
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	"slices"
	"strings"
)

func invalidateBasicUserDtoCache(old, current *UserDTO.Full) {
	invalidator := cache.NewBasicUserDtoInvalidator(old, current)
	// TODO handle error
	if err := invalidator.Invalidate(); err != nil {
		log.DB.Error("Failed to invalidate cache", err.Error(), nil)
	}
}

func (m *Manager) ChangeLogin(act *ActionDTO.UserTargeted, newLogin string) *Error.Status {
	log.DB.Info("Changing login of user "+act.TargetUID+"...", nil)

    if err := act.ValidateUIDs(); err != nil {
		log.DB.Error("Failed to change login of user "+act.TargetUID, err.Error(), nil)
        return err
    }

    if err := user.ValidateLogin(newLogin); err != nil {
		log.DB.Error("Failed to change login of user "+act.TargetUID, err.Error(), nil)
        return err
    }

	if err := authz.User.ChangeUserLogin(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := m.GetUserByID(act.TargetUID)
    if err != nil {
        return err
    }

    if user.Login == newLogin {
		errMsg := "Login not changed: Current login and new login are the same"
		log.DB.Error("Failed to change login of user "+act.TargetUID, err.Error(), nil)
        return Error.NewStatusError(errMsg, http.StatusConflict)
    }

    if err := m.checkIfLoginInUse(newLogin); err != nil {
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

	log.DB.Info("Changing login of user "+act.TargetUID+": OK", nil)

	return nil
}

func (m *Manager) ChangePassword(act *ActionDTO.UserTargeted, newPassword string) *Error.Status {
	log.DB.Info("Changing password of user "+act.TargetUID+"...", nil)

    if err := act.ValidateUIDs(); err != nil {
		log.DB.Error("Failed to change password of user "+act.TargetUID, err.Error(), nil)
        return err
    }

    if err := user.ValidatePassword(newPassword); err != nil {
		log.DB.Error("Failed to change password of user "+act.TargetUID, err.Error(), nil)
        return err
    }

	if err := authz.User.ChangeUserPassword(
		act.RequesterUID == act.TargetUID,
		act.RequesterRoles,
	); err != nil {
		return err
	}

    user, err := m.GetUserByID(act.TargetUID)
    if err != nil {
        return err
    }

	hashedPassword, e := hashPassword(newPassword)
    if e != nil {
		log.DB.Error("Failed to change password of user "+act.TargetUID, e.Error(), nil)
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

	log.DB.Info("Changing password of user "+act.TargetUID+": OK", nil)

	return nil
}

func (m *Manager) ChangeRoles(act *ActionDTO.UserTargeted, newRoles []string) *Error.Status {
	log.DB.Info("Changing roles of user "+act.TargetUID+"...", nil)

	main_loop:
	for _, newRole := range newRoles {
		for _, role := range authz.Schema.Roles {
			if role.Name == newRole {
				continue main_loop
			}
		}

		errMsg := "Role '"+newRole+"' doesn't exists"
		log.DB.Error("Failed to change roles of user "+act.TargetUID, errMsg, nil)
		return Error.NewStatusError(errMsg, http.StatusBadRequest)
	}

    if err := act.ValidateUIDs(); err != nil {
		log.DB.Error("Failed to change roles of user "+act.TargetUID, err.Error(), nil)
        return err
    }

	isRequesterAdmin := slices.Contains(act.RequesterRoles, "admin")
	isAdminInNewRoles := slices.Contains(newRoles, "admin")

    if act.TargetUID == act.RequesterUID && isRequesterAdmin && !isAdminInNewRoles {
		errMsg := "Нельзя снять роль администратора с самого себя"
		log.DB.Error("Failed to change roles of user "+act.TargetUID, errMsg, nil)
		return Error.NewStatusError(errMsg, http.StatusForbidden)
    }

    user, err := m.GetUserByID(act.TargetUID)
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

	log.DB.Info("Changing roles of user "+act.TargetUID+": OK", nil)

	return nil
}

func (m *Manager) Activate(tk string) *Error.Status {
	log.DB.Info("Activating user...", nil)

    t, err := token.ParseSingedToken(tk, config.Secret.ActivationTokenPublicKey)
    if err != nil {
        return err
    }

    payload := UserMapper.PayloadFromClaims(t.Claims.(*token.Claims))

    user, err := m.GetUserByID(payload.ID)
    if err != nil {
        return err
    }
    if user.IsActive() {
		errMsg := "User already active"
		log.DB.Error("Failed activate user", errMsg, nil)
        return Error.NewStatusError(errMsg, http.StatusConflict)
    }

    filter := ActionDTO.NewUserTargeted(user.ID, user.ID, user.Roles)

    audit := newAuditDTO(audit.UpdatedOperation, filter, user)
	var updatedUser *UserDTO.Full

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
		log.DB.Error(
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

	log.DB.Info("Activating user: OK", nil)

    return nil
}

