package auth

import "sentinel/packages/models/role"

type rulebook struct {
	SoftDeleteUser         *authorizationRules
	RestoreSoftDeletedUser *authorizationRules
	DropUser               *authorizationRules
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
		Operation:         AuthorizationOperations.None,
		ValidRoles:        role.List[:],
		ForbidModToModOps: false,
	},
	SoftDeleteUser: &authorizationRules{
		Operation:         AuthorizationOperations.SoftDeleteUser,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: true,
	},
	RestoreSoftDeletedUser: &authorizationRules{
		Operation:         AuthorizationOperations.RestoreSoftDeletedUser,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: true,
	},
	DropUser: &authorizationRules{
		Operation:         AuthorizationOperations.DropUser,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: true,
	},
	ChangeUserLogin: &authorizationRules{
		Operation:         AuthorizationOperations.ChangeUserLogin,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: true,
	},
	ChangeUserPassword: &authorizationRules{
		Operation:         AuthorizationOperations.ChangeUserPassword,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: true,
	},
	ChangeUserRole: &authorizationRules{
		Operation:         AuthorizationOperations.ChangeUserRole,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: true,
	},
	GetUserRole: &authorizationRules{
		Operation:         AuthorizationOperations.GetUserRole,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: false,
	},
}
