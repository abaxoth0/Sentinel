package auth

import (
	"fmt"
	"net/http"
	"sentinel/packages/cache"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"
	"slices"
)

type authorizationRules struct {
	// Unique name of operation.
	Operation OperationName

	// Array of permissions which allows to perform this operation.
	// If user has any of these permissions, he can perform this operation. (It's 'OR' not 'AND')
	RequiredPermission []role.Permission
}

// Verifies if a user with a given role can perform an operation on a target with a specified role.
// It checks both user and target roles against the required permissions for the operation defined in the authorizationRules.
// Returns an ExternalError if the user lacks the necessary permissions or attempts forbidden operations, such as
// modifying an admin or performing moderator-to-moderator actions. Returns nil if authorization is successful.
//
// This method authorize operations only in THIS service!
// Operations on other services must be authorized by themselves!
func (authRules authorizationRules) Authorize(userRoleName string, targetRoleName string) *ExternalError.Error {
	cacheKey := fmt.Sprintf("%s[%s->%s]", string(authRules.Operation), userRoleName, targetRoleName)
	cacheOK := "OK"

	if cacheValue, hit := cache.Get(cacheKey); hit {
		if cacheValue == cacheOK {
			return nil
		}

		return ExternalError.New(cacheValue, http.StatusForbidden)
	}

	err := authorize(authRules, userRoleName, targetRoleName)

	cacheValue := cacheOK

	if err != nil {
		cacheValue = err.Error()
	}

	cache.Set(cacheKey, cacheValue)

	return err
}

func authorize(authRules authorizationRules, userRoleName string, targetRoleName string) *ExternalError.Error {
	insufficientPermission := ExternalError.New("Недостаточно прав для выполнения данной операции", http.StatusForbidden)

	userRole, err := role.ParseRole(userRoleName)

	if err != nil {
		return err
	}

	isUserModerator := slices.Contains(userRole.Permissions, role.ModeratorPermission)
	isUserAdmin := slices.Contains(userRole.Permissions, role.AdminPermission)

	canUserCreate := slices.Contains(userRole.Permissions, role.CreatePermission)
	canUserSelfCreate := slices.Contains(userRole.Permissions, role.SelfCreatePermission)

	canUserRead := slices.Contains(userRole.Permissions, role.ReadPermission)
	canUserSelfRead := slices.Contains(userRole.Permissions, role.SelfReadPermission)

	canUserUpdate := slices.Contains(userRole.Permissions, role.UpdatePermission)
	canUserSelfUpdate := slices.Contains(userRole.Permissions, role.SelfUpdatePermission)

	canUserDelete := slices.Contains(userRole.Permissions, role.DeletePermission)
	canUserSelfDelete := slices.Contains(userRole.Permissions, role.SelfDeletePermission)

	isCreateRequired := slices.Contains(authRules.RequiredPermission, role.CreatePermission)
	isSelfCreateRequired := slices.Contains(authRules.RequiredPermission, role.SelfCreatePermission)

	isReadRequired := slices.Contains(authRules.RequiredPermission, role.ReadPermission)
	isSelfReadRequired := slices.Contains(authRules.RequiredPermission, role.SelfReadPermission)

	isUpdateRequired := slices.Contains(authRules.RequiredPermission, role.UpdatePermission)
	isSelfUpdateRequired := slices.Contains(authRules.RequiredPermission, role.SelfUpdatePermission)

	isDeleteRequired := slices.Contains(authRules.RequiredPermission, role.DeletePermission)
	isSelfDeleteRequired := slices.Contains(authRules.RequiredPermission, role.SelfDeletePermission)

	isAdminRequired := slices.Contains(authRules.RequiredPermission, role.AdminPermission)
	isModeratorRequired := slices.Contains(authRules.RequiredPermission, role.ModeratorPermission)

	// All operations which haven't targetRole requires admin rights.
	// (For example: clear cache, drop all soft deleted users)
	if targetRoleName == "none" && !isUserAdmin {
		return insufficientPermission
	}

	targetRole, err := role.ParseRole(targetRoleName)

	if err != nil {
		return err
	}

	isTargetModerator := slices.Contains(targetRole.Permissions, role.ModeratorPermission)
	isTargetAdmin := slices.Contains(targetRole.Permissions, role.AdminPermission)

	if isAdminRequired && !isUserAdmin {
		return insufficientPermission
	}

	if isModeratorRequired && !isUserModerator && !isUserAdmin {
		return insufficientPermission
	}

	if (isDeleteRequired || isSelfDeleteRequired) && isTargetAdmin {
		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	if isUserModerator && isTargetModerator && (isUpdateRequired || isDeleteRequired) {
		return ExternalError.New("Для данной операции запрещено взаимодействие вида \"модератор-модератор\"", http.StatusForbidden)
	}

	if isTargetModerator && !isUserAdmin && (isUpdateRequired || isDeleteRequired) {
		return insufficientPermission
	}

	if isUserAdmin || isUserModerator {
		return nil
	}

	if isCreateRequired && (!canUserCreate) {
		return insufficientPermission
	}

	if isSelfCreateRequired && (!canUserCreate || !canUserSelfCreate) {
		return insufficientPermission
	}

	if isReadRequired && (!canUserRead) {
		return insufficientPermission
	}

	if isSelfReadRequired && (!canUserRead || !canUserSelfRead) {
		return insufficientPermission
	}

	if isUpdateRequired && (!canUserUpdate) {
		return insufficientPermission
	}

	if isSelfUpdateRequired && (!canUserUpdate || !canUserSelfUpdate) {
		return insufficientPermission
	}

	if isDeleteRequired && (!canUserDelete) {
		return insufficientPermission
	}

	if isSelfDeleteRequired && (!canUserDelete || !canUserSelfDelete) {
		return insufficientPermission
	}

	return nil
}
