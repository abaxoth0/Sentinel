package token

import (
	"net/http"
	Error "sentinel/packages/common/errors"

	"github.com/golang-jwt/jwt"
)

var invalidTokenPayload = Error.NewStatusError(
    "Ошибка авторизации (invalid token payload)",
    http.StatusBadRequest,
)

func VerifyClaims(claims jwt.MapClaims) *Error.Status {
	if claims[IdKey] == nil {
		return invalidTokenPayload
    }

	if claims[IssuerKey] == nil {
		return invalidTokenPayload
    }

	if claims[SubjectKey] == nil {
		return invalidTokenPayload
    }

	return nil
}

