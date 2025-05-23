package authorization

import (
	rbac "github.com/StepanAnanin/SentinelRBAC"
)

type action struct {
	SoftDelete         rbac.Action
	Restore            rbac.Action
	Drop               rbac.Action
	DropAllSoftDeleted rbac.Action
	ChangeLogin        rbac.Action
	ChangePassword     rbac.Action
	ChangeRoles        rbac.Action
	GetRoles           rbac.Action
}

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

var Action = func() *action{
    act := new(action)

    act.SoftDelete = registerAction(
        user,
        "soft_delete",
        rbac.DeletePermission,
    )

    act.Restore = registerAction(
        user,
        "restore",
        rbac.DeletePermission|rbac.UpdatePermission,
    )

    act.Drop = registerAction(
        user,
        "drop",
        rbac.DeletePermission,
    )

    act.DropAllSoftDeleted = registerAction(
        user,
        "drop_all_deleted",
        rbac.DeletePermission,
    )

    act.ChangeLogin = registerAction(
        user,
        "change_login",
        rbac.UpdatePermission,
    )

    act.ChangePassword = registerAction(
        user,
        "change_password",
        rbac.UpdatePermission,
    )

    act.ChangeRoles = registerAction(
        user,
        "change_role",
        rbac.UpdatePermission,
    )

    act.GetRoles = registerAction(
        user,
        "get_roles",
        rbac.ReadPermission,
    )

    return act
}()

