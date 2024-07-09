package auth

import (
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"
	"slices"
)

type authorizationRules struct {
	// Unique name of operation.
	Operation OperationName

	// Array of roles that allow to perform this operation.
	ValidRoles []role.Role

	// Forbid moderator to perform operations with another moderator.
	ForbidModToModOps bool
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

	if !slices.Contains(authRules.ValidRoles, userRole) {
		return ExternalError.New("Недостаточно прав для выполнения данной операции", http.StatusForbidden)
	}

	return nil
}
