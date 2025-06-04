package authz

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
)

type user struct {
	//
}

// Used for authorizing user to perform some action
var User = user{}

func (u user) SoftDeleteUser(self bool, roles []string) *Error.Status {
	return authorize(
		util.Ternary(self, softDeleteSelfAction, softDeleteAction),
		userResource,
		roles,
	)
}

func (u user) RestoreUser(roles []string) *Error.Status {
	return authorize(restoreAction, userResource, roles)
}

func (u user) DropUser(roles []string) *Error.Status {
	return authorize(dropAction, userResource, roles)
}

func (u user) DropAllSoftDeletedUsers(roles []string) *Error.Status {
	return authorize(dropAllSoftDeletedAction, userResource, roles)
}

func (u user) ChangeUserLogin(self bool, roles []string) *Error.Status {
	return authorize(
		util.Ternary(self, changeSelfLoginAction, changeLoginAction),
		userResource,
		roles,
	)
}

func (u user) ChangeUserPassword(self bool, roles []string) *Error.Status {
	return authorize(
		util.Ternary(self, changeSelfPasswordAction, changePasswordAction),
		userResource,
		roles,
	)
}

func (u user) ChangeUserRoles(self bool, roles []string) *Error.Status {
	return authorize(
		util.Ternary(self, changeSelfRolesAction, changeRolesAction),
		userResource,
		roles,
	)
}

func (u user) GetUserRoles(roles []string) *Error.Status {
	return authorize(getRolesAction, userResource, roles)
}

func (u user) DropCache(roles []string) *Error.Status {
	return authorize(dropAction, cacheResource, roles)
}

