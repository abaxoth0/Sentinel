package net

type AuthRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponseBody struct {
	Message     string `json:"message"`
	AccessToken string `json:"accessToken"`
}

type MessageResponseBody struct {
	Message string `json:"message"`
}
