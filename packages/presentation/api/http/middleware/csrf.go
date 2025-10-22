package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Constant-time comparison to prevent timing attacks
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// Requires two same CSRF tokens, one must be set in HTTP-Only cookie and
// another one must be provided in X-CSRF-Token header.
// If tokens doesn't match this middleware won't pass this request further.
func DoubleSubmitCSRF(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Method == http.MethodGet || ctx.Request().Method == http.MethodHead {
			return next(ctx)
		}

		headerToken := ctx.Request().Header.Get("X-CSRF-Token")
		if headerToken == "" {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"CSRF token is missing in the request header",
			)
		}

		cookie, err := ctx.Cookie("_csrf")
		if err != nil {
			if err == http.ErrNoCookie {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					"CSRF cookie is missing",
				)
			} else {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					"Failed to get CSRF cookie",
				)
			}
		}

		if !secureCompare(headerToken, cookie.Value) {
			return echo.NewHTTPError(
				http.StatusForbidden,
				"CSRF tokens mismatch",
			)
		}

		return next(ctx)
	}
}
