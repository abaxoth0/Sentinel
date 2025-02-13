package token

import (
	"net/http"
	Error "sentinel/packages/errs"
	"sentinel/packages/infrastructure/config"
	"sentinel/packages/util"

	"github.com/golang-jwt/jwt"
)

func generateAccessTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.AccessTokenTTL)
}

func generateRefreshTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.RefreshTokenTTL)
}

func VerifyClaims(claims jwt.MapClaims) *Error.Status {
	if claims[IdKey] == nil {
		return Error.NewStatusError("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	if claims[IssuerKey] == nil {
		return Error.NewStatusError("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	if claims[SubjectKey] == nil {
		return Error.NewStatusError("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	return nil
}
