package router

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/DB"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	controller "sentinel/packages/presentation/api/http/controllers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func setTokenRefreshRequiredHeader(ctx echo.Context) {
	ctx.Response().Header().Set("X-Token-Refresh-Required", "true")
}

// Without this user won't be able to refresh auth tokens, login or logout on desync
var skipSyncCheckEndpoints map[string]bool = map[string]bool{
	http.MethodPut + "/auth": true, // Refresh auth tokens endpoint
	http.MethodPost + "/auth": true, // Login endpoint
	http.MethodDelete + "/auth": true, // Logout endpoint
}

func preventUserDesync(next echo.HandlerFunc) echo.HandlerFunc {
	return func (ctx echo.Context) error {
		req := ctx.Request()

		if skipSyncCheckEndpoints[req.Method+req.URL.Path] {
			return next(ctx)
		}

		refreshToken, err := controller.GetRefreshToken(ctx)
		if err != nil {
			return next(ctx)
		}

		payload, err := UserMapper.PayloadFromClaims(refreshToken.Claims.(jwt.MapClaims))
		if err != nil {
			setTokenRefreshRequiredHeader(ctx)
			return echo.NewHTTPError(
				http.StatusUnauthorized,
				"Failed to check user data synchronization: " + err.Error(),
			)
		}

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

		return next(ctx)
	}
}

