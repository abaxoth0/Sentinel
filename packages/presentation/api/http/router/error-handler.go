package router

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	responsebody "sentinel/packages/presentation/data/response"

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

	ctx.JSON(code, responsebody.Error{
		Error: statusText,
		Message: message,
	})
}

