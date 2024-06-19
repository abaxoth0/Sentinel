package auth

import "sentinel/packages/models/role"

type rulebook struct {
	// No rules
	Clear *authorizationRules
	// User soft delete operation rules
	SoftDeleteUser *authorizationRules
	// RestoreSoftDeletedUser soft deleted user
	RestoreSoftDeletedUser *authorizationRules
	// Hard delete user
	DropUser *authorizationRules
	// Change user email
	ChangeUserEmail *authorizationRules
	// Change user password
	ChangeUserPassword *authorizationRules
	// Change user role
	ChangeUserRole *authorizationRules
}

// Used for authorization
var Rulebook = &rulebook{
	Clear: &authorizationRules{
		Operation:                AuthorizationOperations.None,
		ValidRoles:               role.List[:],
		SkipRoleValidationOnSelf: false,
		ForbidModToModOps:        false,
	},
	SoftDeleteUser: &authorizationRules{
		Operation:                AuthorizationOperations.SoftDeleteUser,
		ValidRoles:               []role.Role{role.Moderator, role.Administrator},
		SkipRoleValidationOnSelf: true,
		ForbidModToModOps:        true,
	},
	RestoreSoftDeletedUser: &authorizationRules{
		Operation:                AuthorizationOperations.RestoreSoftDeletedUser,
		ValidRoles:               []role.Role{role.Moderator, role.Administrator},
		SkipRoleValidationOnSelf: false,
		ForbidModToModOps:        true,
	},
	DropUser: &authorizationRules{
		Operation:                AuthorizationOperations.DropUser,
		ValidRoles:               []role.Role{role.Moderator, role.Administrator},
		SkipRoleValidationOnSelf: false,
		ForbidModToModOps:        true,
	},
	ChangeUserEmail: &authorizationRules{
		Operation:                AuthorizationOperations.ChangeUserEmail,
		ValidRoles:               []role.Role{role.Moderator, role.Administrator},
		SkipRoleValidationOnSelf: true,
		ForbidModToModOps:        true,
	},
	ChangeUserPassword: &authorizationRules{
		Operation:                AuthorizationOperations.ChangeUserPassword,
		ValidRoles:               []role.Role{role.Moderator, role.Administrator},
		SkipRoleValidationOnSelf: true,
		ForbidModToModOps:        true,
	},
	ChangeUserRole: &authorizationRules{
		Operation:                AuthorizationOperations.ChangeUserRole,
		ValidRoles:               []role.Role{role.Moderator, role.Administrator},
		SkipRoleValidationOnSelf: false,
		ForbidModToModOps:        true,
	},
}
