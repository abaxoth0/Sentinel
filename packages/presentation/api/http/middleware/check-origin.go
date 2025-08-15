package middleware

import (
	"net/http"
	"sentinel/packages/common/config"
	"sentinel/packages/presentation/api/http/request"
	"slices"

	"github.com/labstack/echo/v4"
)

// Used to prevent request forgery attacks
func CheckOrigin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		req := ctx.Request()

		if req.Method == http.MethodGet || req.Method == http.MethodHead {
			return next(ctx)
		}

		origin := req.Header.Get("Origin")

		if origin != "" && !slices.Contains(config.HTTP.AllowedOrigins, origin) {
			log.Error("Invalid request origin", "Origin isn't allowed", request.GetMetadata(ctx))
			return echo.NewHTTPError(
				http.StatusForbidden,
				"Invalid origin",
			)
		}

		return next(ctx)
	}
}

