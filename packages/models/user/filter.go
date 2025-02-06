package user

import (
	"sentinel/packages/entities"
	"sentinel/packages/models/token"
	"sentinel/packages/util"
	"strings"

	Error "sentinel/packages/errs"

	"github.com/golang-jwt/jwt"
)

const NoTarget string = "no-targeted-user"

func NewFilterFromClaims(targetUID string, claims jwt.MapClaims) (*entities.UserFilter, *Error.HTTP) {
	var r *entities.UserFilter

	if err := token.VerifyClaims(claims); err != nil {
		return r, err
	}

	return &entities.UserFilter{
		TargetUID:      util.Ternary(targetUID == NoTarget, NoTarget, targetUID),
		RequesterUID:   claims[token.IdKey].(string),
		RequesterRoles: strings.Split((claims[token.SubjectKey].(string)), ","),
	}, nil
}

