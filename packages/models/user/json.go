package user

import "sentinel/packages/models/role"

type Payload struct {
	ID    string    `json:"id"`
	Login string    `json:"login"`
	Role  role.Role `json:"role"`
}
