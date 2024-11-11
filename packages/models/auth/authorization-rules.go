package auth

import (
	"fmt"
	"net/http"
	"sentinel/packages/cache"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"
)

type authorizationRules struct {
	// Unique name of operation.
	Operation OperationName

	// Array of permissions which allows to perform this operation.
	// If user has any of these permissions, he can perform this operation. (It's 'OR' not 'AND')
	RequiredPermission []role.PermissionTag
}

// Verifies if a user with a given role can perform an operation on a target with a specified role.
// It checks both user and target roles against the required permissions for the operation defined in the authorizationRules.
// Returns an ExternalError if the user lacks the necessary permissions or attempts forbidden operations, such as
// modifying an admin or performing moderator-to-moderator actions. Returns nil if authorization is successful.
//
// This method authorize operations only in THIS service!
// Operations on other services must be authorized by themselves!
func (rules authorizationRules) Authorize(userRoleName string, targetRoleName string) *ExternalError.Error {
	cacheKey := fmt.Sprintf("%s[%s->%s]", string(rules.Operation), userRoleName, targetRoleName)
	cacheOK := "OK"

	if cacheValue, hit := cache.Get(cacheKey); hit {
		if cacheValue == cacheOK {
			return nil
		}

		return ExternalError.New(cacheValue, http.StatusForbidden)
	}

	err := authorize(rules, userRoleName, targetRoleName)

	cacheValue := cacheOK

	if err != nil {
		cacheValue = err.Error()
	}

	cache.Set(cacheKey, cacheValue)

	return err
}

func authorize(rules authorizationRules, userRoleName string, targetRoleName string) *ExternalError.Error {
	userRole, err := role.ParseRole(userRoleName)

	if err != nil {
		return err
	}

	userPermissions, requiredPermissions := role.GetPermissions(rules.RequiredPermission, userRole)

	// All operations which has targetRoleName == role.NoneRole, but need authorization requires admin rights.
	// (For example: drop all soft deleted users)
	if targetRoleName == role.NoneRole && !userPermissions.Admin {
		return role.InsufficientPermission
	}

	targetRole, err := role.ParseRole(targetRoleName)

	if err != nil {
		return err
	}

	return role.VerifyPermissions(requiredPermissions, userPermissions, targetRole)
}
