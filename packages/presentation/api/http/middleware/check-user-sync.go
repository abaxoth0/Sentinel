package middleware

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/DB"
	SharedController "sentinel/packages/presentation/api/http/controllers/shared"
	"sentinel/packages/presentation/api/http/request"

	"github.com/labstack/echo/v4"
)

func setTokenRefreshRequiredHeader(ctx echo.Context) {
	ctx.Response().Header().Set("X-Token-Refresh-Required", "true")
}

// IMPORTANT: Works only if route\group was secured via 'secure' middleware.
func CheckUserSync(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		reqMeta := request.GetMetadata(ctx)

		if secured := ctx.Get("Secured"); secured == nil || !secured.(bool) {
			log.Panic(
				"Failed to check user data synchronization",
				"Invalid usage of preventUserDesync middleware: route/group must be secured via 'secure' middleware",
				reqMeta,
			)
		}

		log.Trace("Checking if user desynced...", reqMeta)

		payload := SharedController.GetUserPayload(ctx)

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

		log.Trace("Checking if user desynced: OK", reqMeta)

		return next(ctx)
	}
}
