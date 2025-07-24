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

func (u user) SearchUsers(roles []string) *Error.Status {
	return authorize(searchUsersAction, userResource, roles)
}

func (u user) Logout(roles []string) *Error.Status {
	return authorize(logoutUserAction, userResource, roles)
}

func (u user) GetUserSession(self bool, roles []string) *Error.Status {
	return authorize(
		util.Ternary(self, getSelfSessionAction, getSessionAction),
		userResource,
		roles,
	)
}

func (u user) DropCache(roles []string) *Error.Status {
	return authorize(dropAction, cacheResource, roles)
}

func (u user) AccessAPIDocs(roles []string) *Error.Status {
	return authorize(accessAPIDocs, docsResource, roles)
}

func (u user) GetSessionLocation(roles []string) *Error.Status {
	return authorize(getSessionLocation, userResource, roles)
}

func (u user) DeleteLocation(roles []string) *Error.Status {
	return authorize(getSessionLocation, userResource, roles)
}

func (u user) GetUser(self bool, roles []string) *Error.Status {
	return authorize(
		util.Ternary(self, getSelf, getUser),
		userResource,
		roles,
	)
}
