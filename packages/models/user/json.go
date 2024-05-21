package user

type Payload struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
