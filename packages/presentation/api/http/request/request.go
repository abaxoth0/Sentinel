package request

import (
	"errors"
	"fmt"
	"sentinel/packages/common/logger"

	"github.com/labstack/echo/v4"
)

// TODO Do the same for error?
const metaKey = "req_meta"

// This middleware must be applied to the router
// for the all functions in this package to work correctly.
func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func (ctx echo.Context) error {
		req := ctx.Request()

		ctx.Set(
			metaKey,
			logger.Meta{
				"addr": req.RemoteAddr,
				"method": req.Method,
				"path": req.URL.Path,
				"user_agent": req.UserAgent(),
			},
		)

		return next(ctx)
	}
}

// Retrieves log metadata from the context.
// Won't work if request.Middleware isn't applied.
func GetLogMeta(ctx echo.Context) (logger.Meta, error) {
	switch m := ctx.Get(metaKey).(type) {
	case logger.Meta:
		return m, nil
	case nil:
		return nil, errors.New("Request meta wasn't set (check if middleware applied correctly)")
	default:
		return nil, fmt.Errorf("Request meta has invalid type. Expected logger.Meta, but got %T", m)
	}
}

