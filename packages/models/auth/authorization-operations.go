package auth

// Unique name of operation.
type OperationName string

type authorizationOps struct {
	None                   OperationName
	SoftDeleteUser         OperationName
	RestoreSoftDeletedUser OperationName
	DropUser               OperationName
	ChangeUserLogin        OperationName
	ChangeUserPassword     OperationName
	ChangeUserRole         OperationName
	DropCache              OperationName
}

var AuthorizationOperations = &authorizationOps{
	None:                   "none",
	SoftDeleteUser:         "soft_delete_user",
	RestoreSoftDeletedUser: "restore_soft_deleted_user",
	DropUser:               "drop_user",
	ChangeUserLogin:        "change_user_login",
	ChangeUserPassword:     "change_user_password",
	ChangeUserRole:         "change_user_role",
	DropCache:              "drop_cache",
}
