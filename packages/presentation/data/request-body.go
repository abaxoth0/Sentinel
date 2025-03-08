package datamodel

type AuthRequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UidBody struct {
	UID string `json:"uid"`
}

type LoginBody struct {
	Login string `json:"login"`
}

type UidAndLoginBody struct {
	UID   string `json:"uid"`
	Login string `json:"login"`
}

type UidAndPasswordBody struct {
	UID      string `json:"uid"`
	Password string `json:"password"`
}

type UidAndRolesBody struct {
	UID   string   `json:"uid"`
	Roles []string `json:"roles"`
}

