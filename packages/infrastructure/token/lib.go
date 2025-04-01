package token

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/config"
	"sentinel/packages/common/util"

	"github.com/golang-jwt/jwt"
)

func generateAccessTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.AccessTokenTTL)
}

func generateRefreshTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.RefreshTokenTTL)
}

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
