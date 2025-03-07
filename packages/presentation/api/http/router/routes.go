package router

import (

	"github.com/labstack/echo/v4"
)

func Create() *echo.Echo {
	router := echo.New()

    router.HideBanner = true
    router.HidePort = true

    router.HTTPErrorHandler = handleHttpError

	return router
}
