package router

import (
	"fmt"
	"net/http"
	Error "sentinel/packages/common/errors"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	responsebody "sentinel/packages/presentation/data/response"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
)

func handleHttpError(err error, ctx echo.Context) {
    if ctx.Response().Committed {
        return
    }

    // TODO check this out
    // ctx.Response().Before()
    // ctx.Response().After()
    code := http.StatusInternalServerError
    message := "Internal Server Error"

    if e, is := err.(*echo.HTTPError); is {
        code = e.Code
        message = e.Message.(string)
    }

    statusText := Error.StatusText(code)

	if code == Error.SessionRevoked {
		ctx.Response().Header().Set("X-Session-Revoked", "true")
		if authCookie, err := controller.GetAuthCookie(ctx); err == nil {
			controller.DeleteCookie(ctx, authCookie)
		}
	}

	reqMeta := request.GetMetadata(ctx)

    controller.Logger.Error(message, statusText, reqMeta)

	// if server error
	if code >= 500 {
		if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func (scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)

				if httpErr, ok := err.(*echo.HTTPError); ok {
					scope.SetExtra("details", httpErr.Message)
					if httpErr.Internal != nil {
						hub.CaptureException(httpErr.Internal)
					} else {
						hub.CaptureMessage(fmt.Sprintf("%v", httpErr))
					}
				} else {
					hub.CaptureException(err)
				}
			})
		}
	}

	ctx.JSON(code, responsebody.Error{
		Error: statusText,
		Message: message,
	})
}

