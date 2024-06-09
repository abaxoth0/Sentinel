package auth

import (
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"
)

// Argument is user role. Before calling this function ensure, that role is valid.
// Must return true and empty string if OK, otherwise - false and error message.
type additionalConditionFunc func(role.Role) (bool, string)

type authorizationRules struct {
	// Unique name of operation.
	Operation OperationName
	// Array of roles that allow to perform this operation.
	ValidRoles []role.Role
	// If true, then role search in `ValidRoles` will be skiped,
	// but only if user performs operation on himself.
	// (examples: user want to change email, password or even delete his profile)
	SkipRoleValidationOnSelf bool
	// Forbid moderator to perform operations with another moderator.
	ForbidModToModOps bool
	// Needed for some operations, if ok should return true and empty string,
	// otherwise should return false and error message.
	//
	// If this property is not needed then set it to `unspecifiedAdditionalCondition`.
	AdditionCondition additionalConditionFunc
}

// Returns true if role is sufficient to perform this operation, false otherwise.
//
// Before using this method ensure that role is valid via "Verify" method of Role type (role.Role).
// (Better to do this inside of controller)
func (authRules authorizationRules) Authorize(userRole role.Role) *ExternalError.Error {
	// Is Moderator-Moderator operation forbidden
	if authRules.ForbidModToModOps && userRole == role.Moderator {
		return ExternalError.New("Для данной операции запрещено взаимодействие вида \"модератор-модератор\"", http.StatusForbidden)
	}

	if !authRules.SkipRoleValidationOnSelf {
		found := false

		for _, validRole := range authRules.ValidRoles {
			if validRole == userRole {
				found = true
				break
			}
		}

		if !found {
			return ExternalError.New("Недостаточно прав для выполнения данной операции", http.StatusForbidden)
		}
	}

	if ok, message := authRules.AdditionCondition(userRole); !ok {
		return ExternalError.New(message, http.StatusForbidden)
	}

	return nil
}

func softDeleteUserAdditionalCondition(userRole role.Role) (bool, string) {
	// TODO FIX: Now work incorrect (userRole is requester role, but must be target role)
	// if userRole == role.Administrator {
	// 	return false, "Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)"
	// }

	return true, ""
}

func unspecifiedAdditionalCondition(_ role.Role) (bool, string) {
	return true, ""
}
