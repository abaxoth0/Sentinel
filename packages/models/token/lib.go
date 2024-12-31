package token

import (
	"net/http"
	"sentinel/packages/config"
	Error "sentinel/packages/errs"
	"sentinel/packages/util"

	"github.com/golang-jwt/jwt"
)

func generateAccessTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.AccessTokenTTL)
}

func generateRefreshTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.RefreshTokenTTL)
}

func verifyClaims(claims jwt.MapClaims) *Error.HTTP {
	if claims[IdKey] == nil {
		return Error.NewHTTP("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	if claims[IssuerKey] == nil {
		return Error.NewHTTP("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	if claims[SubjectKey] == nil {
		return Error.NewHTTP("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	return nil
}
