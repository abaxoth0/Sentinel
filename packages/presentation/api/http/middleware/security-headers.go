package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func SecurityHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		res := ctx.Response()

		// Prevent HTTPS downgrade attacks
		res.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Block MIME-type sniffing
		res.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking via iframes
		res.Header().Set("X-Frame-Options", "DENY")

		// XSS Protection (legacy browsers)
		res.Header().Set("X-XSS-Protection", "1; mode=block")

		// Special handling for Swagger UI
		if strings.HasPrefix(ctx.Path(), "/docs") {
			res.Header().Set("Content-Security-Policy",
				"default-src 'self';"+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval';"+ // Required for Swagger
					"style-src 'self' 'unsafe-inline';"+ // Required for inline styles
					"img-src 'self' data:;"+ // Allow data URIs for images
					"font-src 'self';"+
					"connect-src 'self';"+ // For API requests
					"frame-ancestors 'none';"+
					"form-action 'self';"+
					"base-uri 'self';")
		} else {
			// Mitigate XSS and data injection
			res.Header().Set("Content-Security-Policy",
				"default-src 'none';"+
					"script-src 'none'; "+
					"frame-ancestors 'none'; "+
					"form-action 'none'; "+
					"base-uri 'none'")
		}

		// Control browser feature access
		res.Header().Set("Permissions-Policy",
			"accelerometer=(), "+
				"camera=(), "+
				"geolocation=(), "+
				"microphone=(), "+
				"usb=()")

		res.Header().Set("Referrer-Policy", "no-referrer")

		return next(ctx)
	}
}
