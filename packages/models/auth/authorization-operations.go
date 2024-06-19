package auth

// Unique name of operation.
type OperationName string

type authorizationOps struct {
	None                   OperationName
	SoftDeleteUser         OperationName
	RestoreSoftDeletedUser OperationName
	DropUser               OperationName
	ChangeUserEmail        OperationName
	ChangeUserPassword     OperationName
	ChangeUserRole         OperationName
}

var AuthorizationOperations = &authorizationOps{
	None:                   "none",
	SoftDeleteUser:         "soft_delete_user",
	RestoreSoftDeletedUser: "restore_soft_deleted_user",
	DropUser:               "drop_user",
	ChangeUserEmail:        "change_user_email",
	ChangeUserPassword:     "change_user_password",
	ChangeUserRole:         "change_user_role",
}
