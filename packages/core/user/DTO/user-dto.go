package userdto

type Indexed struct {
    ID        string   `json:"id"`
	Login     string   `json:"login"`
	Password  string   `json:"password"`
	Roles     []string `json:"roles"`
	DeletedAt int64    `json:"deletedAt"`
}

type Payload struct {
	ID    string   `json:"id"`
	Login string   `json:"login"`
	Roles []string `json:"roles"`
}

type Filter struct {
	TargetUID      string
	RequesterUID   string
	RequesterRoles []string
}

