package token

import (
	"net/http"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/util"

	"github.com/golang-jwt/jwt"
)

func generateAccessTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.AccessTokenTTL)
}

func generateRefreshTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.RefreshTokenTTL)
}

func verifyClaims(claims jwt.MapClaims) *ExternalError.Error {
	if claims[IdKey] == nil {
		return ExternalError.New("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	if claims[IssuerKey] == nil {
		return ExternalError.New("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	if claims[SubjectKey] == nil {
		return ExternalError.New("Ошибка авторизации (invalid token payload)", http.StatusBadRequest)
	}

	return nil
}
