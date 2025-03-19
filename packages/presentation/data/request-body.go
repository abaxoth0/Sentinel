package datamodel

type UidGetter interface {
    GetUID() string
}

type AuthRequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UidBody struct {
	UID string `json:"uid"`
}

func (body *UidBody) GetUID() string {
    return body.UID
}

type LoginBody struct {
	Login string `json:"login"`
}

type UidAndLoginBody struct {
	UID   string `json:"uid"`
	Login string `json:"login"`
}

func (body *UidAndLoginBody) GetUID() string {
    return body.UID
}

type UidAndPasswordBody struct {
	UID      string `json:"uid"`
	Password string `json:"password"`
}

func (body *UidAndPasswordBody) GetUID() string {
    return body.UID
}

type UidAndRolesBody struct {
	UID   string   `json:"uid"`
	Roles []string `json:"roles"`
}

func (body *UidAndRolesBody) GetUID() string {
    return body.UID
}

