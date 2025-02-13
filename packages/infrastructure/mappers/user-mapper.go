package usermapper

import (
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/util"
	"strings"

	Error "sentinel/packages/errs"

	"github.com/golang-jwt/jwt"
)

const NoTarget string = "no-targeted-user"

func FilterDTOFromClaims(targetUID string, claims jwt.MapClaims) (*UserDTO.Filter, *Error.Status) {
	var r *UserDTO.Filter

	if err := token.VerifyClaims(claims); err != nil {
		return r, err
	}

	return &UserDTO.Filter{
		TargetUID:      util.Ternary(targetUID == NoTarget, NoTarget, targetUID),
		RequesterUID:   claims[token.IdKey].(string),
		RequesterRoles: strings.Split((claims[token.SubjectKey].(string)), ","),
	}, nil
}

// IMPORTANT: Use this function only if token is valid.
func PayloadFromClaims(claims jwt.MapClaims) (*UserDTO.Payload, *Error.Status) {
	var r *UserDTO.Payload

	if err := token.VerifyClaims(claims); err != nil {
		return r, err
	}

	return &UserDTO.Payload{
		ID:    claims[token.IdKey].(string),
		Login: claims[token.IssuerKey].(string),
		Roles: strings.Split((claims[token.SubjectKey].(string)), ","),
	}, nil
}
