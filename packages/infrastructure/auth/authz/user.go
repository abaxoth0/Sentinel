package authz

import (
	Error "sentinel/packages/common/errors"
)

type user struct {
	//
}

// Used for users authorization
var User = user{}

func (u user) SoftDeleteUser(self bool, roles []string) *Error.Status {
	if self {
		return authorize(&userSoftDeleteSelfContext, roles)
	}
	return authorize(&userSoftDeleteUserContext, roles)
}

func (u user) RestoreUser(roles []string) *Error.Status {
	return authorize(&userRestoreUserContext, roles)
}

func (u user) DropUser(roles []string) *Error.Status {
	return authorize(&userDropUserContext, roles)
}

func (u user) DropAllSoftDeletedUsers(roles []string) *Error.Status {
	return authorize(&userDropAllSoftDeletedUsersContext, roles)
}

func (u user) ChangeUserLogin(self bool, roles []string) *Error.Status {
	if self {
		return authorize(&userChangeSelfLoginContext, roles)
	}
	return authorize(&userChangeUserLoginContext, roles)
}

func (u user) ChangeUserPassword(self bool, roles []string) *Error.Status {
	if self {
		return authorize(&userChangeSelfPasswordContext, roles)
	}
	return authorize(&userChangeUserPasswordContext, roles)
}

func (u user) ChangeUserRoles(self bool, roles []string) *Error.Status {
	if self {
		return authorize(&userChangeSelfRolesContext, roles)
	}
	return authorize(&userChangeUserRolesContext, roles)
}

func (u user) GetUserRoles(roles []string) *Error.Status {
	return authorize(&userGetUserRolesContext, roles)
}

func (u user) SearchUsers(roles []string) *Error.Status {
	return authorize(&userSearchUsersContext, roles)
}

func (u user) Logout(roles []string) *Error.Status {
	return authorize(&userLogoutUserContext, roles)
}

func (u user) GetUserSession(self bool, roles []string) *Error.Status {
	if self {
		return authorize(&userGetSelfSessionContext, roles)
	}
	return authorize(&userGetSessionContext, roles)
}

func (u user) DropCache(roles []string) *Error.Status {
	return authorize(&userDropCacheContext, roles)
}

func (u user) AccessAPIDocs(roles []string) *Error.Status {
	return authorize(&userAccessAPIDocsContext, roles)
}

func (u user) GetSessionLocation(roles []string) *Error.Status {
	return authorize(&userGetSessionLocationContext, roles)
}

func (u user) DeleteLocation(roles []string) *Error.Status {
	return authorize(&userGetSessionLocationContext, roles)
}

func (u user) OAuthIntrospect(roles []string) *Error.Status {
	return authorize(&userIntrospectOAuthTokenContext, roles)
}

