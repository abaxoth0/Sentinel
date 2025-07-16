package authz

import (
	rbac "github.com/StepanAnanin/SentinelRBAC"
)

func registerAction(
    entity rbac.Entity,
    action string,
    permissions rbac.Permissions,
) rbac.Action{
    act, err := entity.NewAction(action, permissions)
    if err != nil {
        panic(err)
    }

    return act
}

var softDeleteAction = registerAction(
	userEntity,
	"soft_delete",
	rbac.DeletePermission,
)

var softDeleteSelfAction = registerAction(
	userEntity,
	"soft_delete_self",
	rbac.SelfDeletePermission,
)

var restoreAction = registerAction(
	userEntity,
	"restore",
	rbac.DeletePermission|rbac.UpdatePermission,
)

var dropAction = registerAction(
	userEntity,
	"drop",
	rbac.DeletePermission,
)

var dropAllSoftDeletedAction = registerAction(
	userEntity,
	"drop_all_deleted",
	rbac.DeletePermission,
)

var changeLoginAction = registerAction(
	userEntity,
	"change_login",
	rbac.UpdatePermission,
)

var changeSelfLoginAction = registerAction(
	userEntity,
	"change_self_login",
	rbac.SelfUpdatePermission,
)

var changePasswordAction = registerAction(
	userEntity,
	"change_password",
	rbac.UpdatePermission,
)

var changeSelfPasswordAction = registerAction(
	userEntity,
	"change_self_password",
	rbac.SelfUpdatePermission,
)

var changeRolesAction = registerAction(
	userEntity,
	"change_roles",
	rbac.UpdatePermission,
)

var changeSelfRolesAction = registerAction(
	userEntity,
	"change_self_roles",
	rbac.SelfUpdatePermission,
)

var getRolesAction = registerAction(
	userEntity,
	"get_roles",
	rbac.ReadPermission,
)

var searchUsersAction = registerAction(
	userEntity,
	"search_users",
	rbac.ReadPermission,
)

var logoutUserAction = registerAction(
	userEntity,
	"logout_user",
	rbac.DeletePermission, // Delete cuz logging out soft deletes session
)

var getSessionAction = registerAction(
	userEntity,
	"get_session",
	rbac.ReadPermission,
)

var getSelfSessionAction = registerAction(
	userEntity,
	"get_self_session",
	rbac.SelfReadPermission,
)

var accessAPIDocs = registerAction(
	userEntity,
	"access_api_docs",
	rbac.ReadPermission,
)

var getSessionLocation = registerAction(
	userEntity,
	"get_session_location",
	rbac.ReadPermission,
)

var deleteLocation = registerAction(
	userEntity,
	"delete_location",
	rbac.DeletePermission,
)

