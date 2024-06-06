package auth

// Unique name of operation.
type OperationName string

type authorizationOps struct {
	None                   OperationName
	SoftDeleteUser         OperationName
	RestoreSoftDeletedUser OperationName
}

var AuthorizationOperations = &authorizationOps{
	None:                   "none",
	SoftDeleteUser:         "soft_delete_user",
	RestoreSoftDeletedUser: "restore_soft_deleted_user",
}
