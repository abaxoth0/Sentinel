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
		RequiredPermission: []role.Permission{},
	},
	SoftDeleteUser: &authorizationRules{
		Operation:          AuthorizationOperations.SoftDeleteUser,
		RequiredPermission: []role.Permission{role.SelfDeletePermission, role.AdminPermission, role.ModeratorPermission},
	},
	RestoreSoftDeletedUser: &authorizationRules{
		Operation:          AuthorizationOperations.RestoreSoftDeletedUser,
		RequiredPermission: []role.Permission{role.SelfUpdatePermission, role.AdminPermission, role.ModeratorPermission},
	},
	DropUser: &authorizationRules{
		Operation:          AuthorizationOperations.DropUser,
		RequiredPermission: []role.Permission{role.DeletePermission, role.AdminPermission, role.ModeratorPermission},
	},
	DropAllDeletedUsers: &authorizationRules{
		Operation:          AuthorizationOperations.DropAllDeletedUsers,
		RequiredPermission: []role.Permission{role.DeletePermission, role.AdminPermission},
	},
	ChangeUserLogin: &authorizationRules{
		Operation:          AuthorizationOperations.ChangeUserLogin,
		RequiredPermission: []role.Permission{role.SelfUpdatePermission, role.AdminPermission, role.ModeratorPermission},
	},
	ChangeUserPassword: &authorizationRules{
		Operation:          AuthorizationOperations.ChangeUserPassword,
		RequiredPermission: []role.Permission{role.SelfUpdatePermission, role.AdminPermission, role.ModeratorPermission},
	},
	ChangeUserRole: &authorizationRules{
		Operation:          AuthorizationOperations.ChangeUserRole,
		RequiredPermission: []role.Permission{role.SelfUpdatePermission, role.AdminPermission, role.ModeratorPermission},
	},
	GetUserRole: &authorizationRules{
		Operation:          AuthorizationOperations.GetUserRole,
		RequiredPermission: []role.Permission{role.SelfReadPermission, role.AdminPermission, role.ModeratorPermission},
	},
}
