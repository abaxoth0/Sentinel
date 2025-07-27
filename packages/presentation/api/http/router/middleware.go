package router

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	"strings"

	"github.com/labstack/echo/v4"
)

var middlewareLogger = logger.NewSource("MIDDLEWARE", logger.Default)

var invalidAuthorizationHeaderFormat = echo.NewHTTPError(
    http.StatusUnauthorized,
	"Authorization header has invalid format. Expected token bearer format. ('Bearer <token>')",
)

// Allows access only for authenticated users.
func secure(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		reqMeta := request.GetMetadata(ctx)

		middlewareLogger.Debug("Route "+ctx.Request().Method+" "+ctx.Path()+" is secured", reqMeta)
		middlewareLogger.Trace("Extracting access token from the request...", reqMeta)

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
			return controller.ConvertErrorStatusToHTTP(err)
		}

		payload := UserMapper.PayloadFromClaims(accessToken.Claims.(*token.Claims))

		act := ActionDTO.NewUserTargeted(payload.ID, payload.ID, payload.Roles)

		if _, err := DB.Database.GetRevokedSessionByID(act, payload.SessionID); err == nil {
			return controller.ConvertErrorStatusToHTTP(Error.StatusSessionRevoked)
		}

		ctx.Set("access_token", accessToken)
		ctx.Set("access_token_payload", payload)
		ctx.Set("basic_action", &act.Basic)
		ctx.Set("Secured", true)

		middlewareLogger.Trace("Extracting access token from the request: OK", reqMeta)

		return next(ctx)
	}
}

func setTokenRefreshRequiredHeader(ctx echo.Context) {
	ctx.Response().Header().Set("X-Token-Refresh-Required", "true")
}

// Without this user won't be able to refresh auth tokens, login or logout on desync
var skipSyncCheckEndpoints map[string]bool = map[string]bool{
	http.MethodPut + "/auth": true, // Refresh auth tokens endpoint
	http.MethodPost + "/auth": true, // Login endpoint
	http.MethodDelete + "/auth": true, // Logout endpoint
}

// IMPORTANT: Works only if route\group was secured via 'secure' middleware.
func preventUserDesync(next echo.HandlerFunc) echo.HandlerFunc {
	return func (ctx echo.Context) error {
		reqMeta := request.GetMetadata(ctx)

		if secured := ctx.Get("Secured"); secured == nil || !secured.(bool) {
			middlewareLogger.Panic(
				"Failed to check user data synchronization",
				"Invalid usage of preventUserDesync middleware: route/group must be secured via 'secure' middleware",
				reqMeta,
			)
		}

		middlewareLogger.Trace("Checking if user desynced...", reqMeta)

		payload := controller.GetAccessTokenPayload(ctx)

		actualVersion, err := DB.Database.GetUserVersion(payload.ID)
		if err != nil {
			setTokenRefreshRequiredHeader(ctx)
			return echo.NewHTTPError(
				http.StatusUnauthorized,
				"Failed to check user data synchronization. Try to refresh auth tokens.",
			)
		}

		if actualVersion != payload.Version {
			setTokenRefreshRequiredHeader(ctx)
			return echo.NewHTTPError(
				Error.Desync,
				"User data desync: there are a newer version of user, refresh auth tokens to fix this error.",
			)
		}

		middlewareLogger.Trace("Checking if user desynced: OK", reqMeta)

		return next(ctx)
	}
}

