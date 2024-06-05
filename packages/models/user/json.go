package user

import "sentinel/packages/models/role"

type Payload struct {
	ID    string    `json:"id"`
	Email string    `json:"email"`
	Role  role.Role `json:"role"`
}
