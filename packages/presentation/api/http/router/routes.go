package router

import (
	Auth "sentinel/packages/presentation/api/http/controllers/auth"

	"github.com/labstack/echo/v4"
)

func Create() *echo.Echo {
	router := echo.New()

    router.HideBanner = true
    router.HidePort = true

    router.HTTPErrorHandler = handleHttpError
    router.JSONSerializer = serializer{}

    router.GET("/auth", Auth.Verify)
    router.POST("/auth", Auth.Login)
    router.PUT("/auth", Auth.Refresh)
    router.DELETE("/auth", Auth.Logout)

	return router
}
