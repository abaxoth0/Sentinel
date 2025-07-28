package request

import (
	"fmt"
	"net/http"
	"sentinel/packages/common/logger"
	transport "sentinel/packages/presentation/api/http"

	"github.com/labstack/echo/v4"
)

// TODO Do the same for error?
const metaKey = "req_meta"

func newMeta(requestID string, req *http.Request) logger.Meta {
	return logger.Meta{
		"request_id": requestID,
		"addr": req.RemoteAddr,
		"method": req.Method,
		"path": req.URL.RequestURI(),
		"user_agent": req.UserAgent(),
	}
}

// This middleware must be applied to the router
// for the all functions in this package to work correctly.
func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func (ctx echo.Context) error {
		id := ctx.Response().Header().Get("X-Request-Id")
		ctx.Set(metaKey, newMeta(id, ctx.Request()))

		return next(ctx)
	}
}

// Retrieves metadata from the context.
// Will panic if request.Middleware wasn't applied to the router.
func GetMetadata(ctx echo.Context) logger.Meta {
	switch m := ctx.Get(metaKey).(type) {
	case logger.Meta:
		return m
	case nil:
		id := ctx.Response().Header().Get("X-Request-Id")

		transport.Logger.Panic(
			"Failed to get metadata from context",
			"Request meta wasn't set (check if middleware applied correctly)",
			newMeta(id, ctx.Request()),
		)
		return nil
	default:
		id := ctx.Response().Header().Get("X-Request-Id")

		transport.Logger.Panic(
			"Failed to get metadata from context",
			fmt.Sprintf("Request meta has invalid type. Expected logger.Meta, but got %T", m),
			newMeta(id, ctx.Request()),
		)
		return nil
	}
}

