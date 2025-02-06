package entities

type User struct {
	Login    string
	Password string
	Roles    []string
}

type UserFilter struct {
	TargetUID      string
	RequesterUID   string
	RequesterRoles []string
}

type UserPayload struct {
	ID    string   `json:"id"`
	Login string   `json:"login"`
	Roles []string `json:"roles"`
}

