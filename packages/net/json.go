package net

type AuthRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UidBody struct {
	UID string `json:"uid"`
}

type UidAndEmailBody struct {
	UID   string `json:"uid"`
	Email string `json:"email"`
}

type UidAndPasswordBody struct {
	UID      string `json:"uid"`
	Password string `json:"password"`
}

type UidAndRoleBody struct {
	UID  string `json:"uid"`
	Role string `json:"role"`
}

type TokenResponseBody struct {
	Message     string `json:"message"`
	AccessToken string `json:"accessToken"`
}

type MessageResponseBody struct {
	Message string `json:"message"`
}
