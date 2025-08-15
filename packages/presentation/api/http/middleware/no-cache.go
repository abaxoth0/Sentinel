package middleware

import "github.com/labstack/echo/v4"

// Used to prevent sensitive data caching at transport layer
func NoCache(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		res := ctx.Response()

		res.Header().Set("Cache-Control", "no-store, max-age=0")
		res.Header().Set("Pragma", "no-cache")
		res.Header().Set("Expires", "0")

		return next(ctx)
	}
}
