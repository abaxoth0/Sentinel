package auth

import "sentinel/packages/models/role"

type rulebook struct {
	SoftDeleteUser         *authorizationRules
	RestoreSoftDeletedUser *authorizationRules
	DropUser               *authorizationRules
	DropAllDeletedUsers    *authorizationRules
	ChangeUserLogin        *authorizationRules
	ChangeUserPassword     *authorizationRules
	ChangeUserRole         *authorizationRules
	GetUserRole            *authorizationRules
	// No rules
	Clear *authorizationRules
}

// Used for authorization
var Rulebook = &rulebook{
	Clear: &authorizationRules{
		Operation:          AuthorizationOperations.None,
		RequiredPermission: []role.PermissionTag{},
	},
	SoftDeleteUser: &authorizationRules{
		Operation:          AuthorizationOperations.SoftDeleteUser,
		RequiredPermission: []role.PermissionTag{role.SelfDeletePermissionTag, role.AdminPermissionTag, role.ModeratorPermissionTag},
	},
	RestoreSoftDeletedUser: &authorizationRules{
		Operation:          AuthorizationOperations.RestoreSoftDeletedUser,
		RequiredPermission: []role.PermissionTag{role.SelfUpdatePermissionTag, role.AdminPermissionTag, role.ModeratorPermissionTag},
	},
	DropUser: &authorizationRules{
		Operation:          AuthorizationOperations.DropUser,
		RequiredPermission: []role.PermissionTag{role.DeletePermissionTag, role.AdminPermissionTag, role.ModeratorPermissionTag},
	},
	DropAllDeletedUsers: &authorizationRules{
		Operation:          AuthorizationOperations.DropAllDeletedUsers,
		RequiredPermission: []role.PermissionTag{role.DeletePermissionTag, role.AdminPermissionTag},
	},
	ChangeUserLogin: &authorizationRules{
		Operation:          AuthorizationOperations.ChangeUserLogin,
		RequiredPermission: []role.PermissionTag{role.SelfUpdatePermissionTag, role.AdminPermissionTag, role.ModeratorPermissionTag},
	},
	ChangeUserPassword: &authorizationRules{
		Operation:          AuthorizationOperations.ChangeUserPassword,
		RequiredPermission: []role.PermissionTag{role.SelfUpdatePermissionTag, role.AdminPermissionTag, role.ModeratorPermissionTag},
	},
	ChangeUserRole: &authorizationRules{
		Operation:          AuthorizationOperations.ChangeUserRole,
		RequiredPermission: []role.PermissionTag{role.SelfUpdatePermissionTag, role.AdminPermissionTag, role.ModeratorPermissionTag},
	},
	GetUserRole: &authorizationRules{
		Operation:          AuthorizationOperations.GetUserRole,
		RequiredPermission: []role.PermissionTag{role.SelfReadPermissionTag, role.AdminPermissionTag, role.ModeratorPermissionTag},
	},
}
