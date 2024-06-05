package auth

import "sentinel/packages/models/role"

type rulebook struct {
	// No rules
	Clear *authorizationRules
	// User soft delete operation rules
	SoftDeleteUser *authorizationRules
}

var Rulebook = &rulebook{
	Clear: &authorizationRules{
		Operation:         AuthorizationOperations.None,
		ValidRoles:        role.List[:],
		ForbidModToModOps: false,
		AdditionCondition: notSpecifiedAdditionalCondition,
	},
	SoftDeleteUser: &authorizationRules{
		Operation:         AuthorizationOperations.SoftDeleteUser,
		ValidRoles:        []role.Role{role.Moderator, role.Administrator},
		ForbidModToModOps: true,
		AdditionCondition: softDeleteUserAdditionalCondition,
	},
}
