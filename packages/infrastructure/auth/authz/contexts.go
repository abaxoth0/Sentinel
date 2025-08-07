package authz

import (
	rbac "github.com/abaxoth0/SentinelRBAC"
)

func newAuthzContext(
    entity *rbac.Entity,
    actionName string,
    requiredPermissions rbac.Permissions,
	resource *rbac.Resource,
) rbac.AuthorizationContext {
	act := rbac.Action(actionName)

	if !entity.HasAction(act) {
		var err error
		act, err = entity.NewAction(actionName, requiredPermissions)
		if err != nil {
			panic(err)
		}
	}

	return rbac.NewAuthorizationContext(entity, act, resource)
}

var (
	userSoftDeleteUserContext 			rbac.AuthorizationContext
	userSoftDeleteSelfContext 			rbac.AuthorizationContext
	userRestoreUserContext 				rbac.AuthorizationContext
	userDropUserContext 				rbac.AuthorizationContext
	userDropAllSoftDeletedUsersContext 	rbac.AuthorizationContext
	userChangeUserLoginContext 			rbac.AuthorizationContext
	userChangeSelfLoginContext 			rbac.AuthorizationContext
	userChangeUserPasswordContext 		rbac.AuthorizationContext
	userChangeSelfPasswordContext 		rbac.AuthorizationContext
	userChangeUserRolesContext 			rbac.AuthorizationContext
	userChangeSelfRolesContext			rbac.AuthorizationContext
	userGetUserRolesContext				rbac.AuthorizationContext
	userSearchUsersContext 				rbac.AuthorizationContext
	userLogoutUserContext 				rbac.AuthorizationContext
	userGetSessionContext 				rbac.AuthorizationContext
	userGetSelfSessionContext 			rbac.AuthorizationContext
	userAccessAPIDocsContext 			rbac.AuthorizationContext
	userGetSessionLocationContext 		rbac.AuthorizationContext
	userDeleteLocationContext 			rbac.AuthorizationContext
	userGetUserContext 					rbac.AuthorizationContext
	userGetSelfContext 					rbac.AuthorizationContext
	userIntrospectOAuthTokenContext 	rbac.AuthorizationContext
	userDropCacheContext 				rbac.AuthorizationContext
)

func initContexts() {
	log.Info("Initializing contexts...", nil)

	userSoftDeleteUserContext = newAuthzContext(
		&userEntity,
		"soft_delete",
		rbac.DeletePermission,
		userResource,
	)

	userSoftDeleteSelfContext = newAuthzContext(
		&userEntity,
		"soft_delete_self",
		rbac.SelfDeletePermission,
		userResource,
	)

	userRestoreUserContext = newAuthzContext(
		&userEntity,
		"restore",
		rbac.DeletePermission|rbac.UpdatePermission,
		userResource,
	)

	userDropUserContext = newAuthzContext(
		&userEntity,
		"drop",
		rbac.DeletePermission,
		userResource,
	)

	userDropAllSoftDeletedUsersContext = newAuthzContext(
		&userEntity,
		"drop_all_deleted",
		rbac.DeletePermission,
		userResource,
	)

	userChangeUserLoginContext = newAuthzContext(
		&userEntity,
		"change_login",
		rbac.UpdatePermission,
		userResource,
	)

	userChangeSelfLoginContext = newAuthzContext(
		&userEntity,
		"change_self_login",
		rbac.SelfUpdatePermission,
		userResource,
	)

	userChangeUserPasswordContext = newAuthzContext(
		&userEntity,
		"change_password",
		rbac.UpdatePermission,
		userResource,
	)

	userChangeSelfPasswordContext = newAuthzContext(
		&userEntity,
		"change_self_password",
		rbac.SelfUpdatePermission,
		userResource,
	)

	userChangeUserRolesContext = newAuthzContext(
		&userEntity,
		"change_roles",
		rbac.UpdatePermission,
		userResource,
	)

	userChangeSelfRolesContext = newAuthzContext(
		&userEntity,
		"change_self_roles",
		rbac.SelfUpdatePermission,
		userResource,
	)

	userGetUserRolesContext = newAuthzContext(
		&userEntity,
		"get_roles",
		rbac.ReadPermission,
		userResource,
	)

	userSearchUsersContext = newAuthzContext(
		&userEntity,
		"search_users",
		rbac.ReadPermission,
		userResource,
	)

	userLogoutUserContext = newAuthzContext(
		&userEntity,
		"logout_user",
		rbac.DeletePermission, // Delete cuz logging out soft deletes session
		userResource,
	)

	userGetSessionContext = newAuthzContext(
		&userEntity,
		"get_session",
		rbac.ReadPermission,
		sessionResource,
	)

	userGetSelfSessionContext = newAuthzContext(
		&userEntity,
		"get_self_session",
		rbac.SelfReadPermission,
		sessionResource,
	)

	userAccessAPIDocsContext = newAuthzContext(
		&userEntity,
		"access_api_docs",
		rbac.ReadPermission,
		docsResource,
	)

	userGetSessionLocationContext = newAuthzContext(
		&userEntity,
		"get_session_location",
		rbac.ReadPermission,
		locationResource,
	)

	userDeleteLocationContext = newAuthzContext(
		&userEntity,
		"delete_location",
		rbac.DeletePermission,
		locationResource,
	)

	userGetUserContext = newAuthzContext(
		&userEntity,
		"get_user",
		rbac.ReadPermission,
		userResource,
	)

	userGetSelfContext = newAuthzContext(
		&userEntity,
		"get_self",
		rbac.SelfReadPermission,
		userResource,
	)

	userIntrospectOAuthTokenContext = newAuthzContext(
		&userEntity,
		"oauth_introspect",
		rbac.ReadPermission,
		oauthTokenResource,
	)

	userDropCacheContext = newAuthzContext(
		&userEntity,
		"drop",
		rbac.DeletePermission,
		cacheResource,
	)

	log.Info("Initializing contexts: OK", nil)
}

