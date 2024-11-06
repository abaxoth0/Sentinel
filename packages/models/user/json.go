package user

type Payload struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Role  string `json:"role"`
}
