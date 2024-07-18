package auth

// Unique name of operation.
type OperationName string

type authorizationOps struct {
	None                   OperationName
	SoftDeleteUser         OperationName
	RestoreSoftDeletedUser OperationName
	DropUser               OperationName
	DropAllDeletedUsers    OperationName
	ChangeUserLogin        OperationName
	ChangeUserPassword     OperationName
	ChangeUserRole         OperationName
	GetUserRole            OperationName
}

var AuthorizationOperations = &authorizationOps{
	None:                   "none",
	SoftDeleteUser:         "soft_delete_user",
	RestoreSoftDeletedUser: "restore_soft_deleted_user",
	DropUser:               "drop_user",
	DropAllDeletedUsers:    "drop_all_deleted_users",
	ChangeUserLogin:        "change_user_login",
	ChangeUserPassword:     "change_user_password",
	ChangeUserRole:         "change_user_role",
	GetUserRole:            "get_user_role",
}
