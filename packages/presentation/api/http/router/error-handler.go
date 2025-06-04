package router

import (
	"net/http"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"

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

    status := http.StatusText(code)

	reqMeta := request.GetMetadata(ctx)

    controller.Logger.Error(message, status, reqMeta)

    ctx.JSON(code, map[string]string{
        "error": status,
        "message": message,
    })
}

