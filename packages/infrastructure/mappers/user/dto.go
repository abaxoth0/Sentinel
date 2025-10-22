package usermapper

import (
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/token"
)

func PayloadFromClaims(claims *token.Claims) *UserDTO.Payload {
	return &UserDTO.Payload{
		ID:        claims.Subject,
		Login:     claims.Login,
		SessionID: claims.ID,
		Roles:     claims.Roles,
		Version:   claims.Version,
		Audience:  claims.Audience,
	}
}
