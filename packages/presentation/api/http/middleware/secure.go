package middleware

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/request"
	"strings"

	"github.com/labstack/echo/v4"
)

var invalidAuthorizationHeaderFormat = echo.NewHTTPError(
    http.StatusUnauthorized,
	"Authorization header has invalid format. Expected token bearer format. ('Bearer <token>')",
)

// Allows access only for authenticated users.
func Secure(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		reqMeta := request.GetMetadata(ctx)

		log.Debug("Route "+ctx.Request().Method+" "+ctx.Path()+" is secured", reqMeta)
		log.Trace("Extracting access token from the request...", reqMeta)

		authHeader := ctx.Request().Header.Get("Authorization")
		if strings.ReplaceAll(authHeader, " ", "") == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized")
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return invalidAuthorizationHeaderFormat
		}

		splitAuthHeader := strings.Split(authHeader, "Bearer ")
		if len(splitAuthHeader) != 2 {
			return invalidAuthorizationHeaderFormat
		}

		accessTokenStr := splitAuthHeader[1]

		accessToken, err := token.ParseSingedToken(accessTokenStr, config.Secret.AccessTokenPublicKey)
		if err != nil {
			return err
		}

		payload := UserMapper.PayloadFromClaims(accessToken.Claims.(*token.Claims))

		act := ActionDTO.NewUserTargeted(payload.ID, payload.ID, payload.Roles)

		if _, err := DB.Database.GetRevokedSessionByID(act, payload.SessionID); err == nil {
			return Error.StatusSessionRevoked
		}

		ctx.Set("access_token", accessToken)
		ctx.Set("user_payload", payload)
		ctx.Set("basic_action", &act.Basic)
		ctx.Set("Secured", true)

		log.Trace("Extracting access token from the request: OK", reqMeta)

		return next(ctx)
	}
}

