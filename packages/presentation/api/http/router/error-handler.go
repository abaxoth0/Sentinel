package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func handleHttpError(err error, ctx echo.Context) {
    ctx.Logger().Error(err)

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

    ctx.JSON(code, map[string]string{
        "error": http.StatusText(code),
        "message": message,
    })
}

