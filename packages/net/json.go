package net

type AuthRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UidBody struct {
	UID string `json:"uid"`
}

type TokenResponseBody struct {
	Message     string `json:"message"`
	AccessToken string `json:"accessToken"`
}

type MessageResponseBody struct {
	Message string `json:"message"`
}
