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

var Action = &action{
	SoftDelete: user.NewAction("soft_delete", &rbac.Permissions{
		Delete: true,
	}),

	Restore: user.NewAction("restore", &rbac.Permissions{
		Delete: true,
		Update: true,
	}),

	Drop: user.NewAction("drop", &rbac.Permissions{
		Delete: true,
	}),

	DropAllSoftDeleted: user.NewAction("drop_all_deleted", &rbac.Permissions{
		Delete: true,
	}),

	ChangeLogin: user.NewAction("change_login", &rbac.Permissions{
		Update: true,
	}),

	ChangePassword: user.NewAction("change_password", &rbac.Permissions{
		Update: true,
	}),

	ChangeRoles: user.NewAction("change_role", &rbac.Permissions{
		Update: true,
	}),

	GetRoles: user.NewAction("get_role", &rbac.Permissions{
		Read: true,
	}),
}
