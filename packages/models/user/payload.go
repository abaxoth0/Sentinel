package user

import (
	"sentinel/packages/entities"
	"sentinel/packages/models/token"
	"strings"

	Error "sentinel/packages/errs"

	"github.com/golang-jwt/jwt"
)

// IMPORTANT: Use this function only if token is valid.
func PayloadFromClaims(claims jwt.MapClaims) (*entities.UserPayload, *Error.HTTP) {
	var r *entities.UserPayload

	if err := token.VerifyClaims(claims); err != nil {
		return r, err
	}

	return &entities.UserPayload{
		ID:    claims[token.IdKey].(string),
		Login: claims[token.IssuerKey].(string),
		Roles: strings.Split((claims[token.SubjectKey].(string)), ","),
	}, nil
}
