package auth

type authorizationOps struct {
	None           string
	SoftDeleteUser string
}

var AuthorizationOperations = &authorizationOps{
	None:           "none",
	SoftDeleteUser: "soft_delete_user",
}
