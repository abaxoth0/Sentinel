package router

import (
	"net/http"
	"sentinel/packages/infrastructure/config"
	Auth "sentinel/packages/presentation/api/http/controllers/auth"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const groupRootPath = ""

func Create() *echo.Echo {
	router := echo.New()

    router.HideBanner = true
    router.HidePort = true

    router.HTTPErrorHandler = handleHttpError
    router.JSONSerializer = serializer{}

    cors := middleware.CORSConfig{
        Skipper:      middleware.DefaultSkipper,
        AllowOrigins: config.HTTP.AllowedOrigins,
        AllowCredentials: true,
        AllowMethods: []string{
            http.MethodGet,
            http.MethodHead,
            http.MethodPut,
            http.MethodPatch,
            http.MethodPost,
            http.MethodDelete,
        },
    }

    router.Use(middleware.CORSWithConfig(cors))
    router.Use(middleware.Recover())
    // router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10_000)))

    if config.Debug.Enabled {
        router.Use(middleware.Logger())
    }

    authGroup := router.Group("/auth")

    authGroup.GET(groupRootPath, Auth.Verify)
    authGroup.POST(groupRootPath, Auth.Login)
    authGroup.PUT(groupRootPath, Auth.Refresh)
    authGroup.DELETE(groupRootPath, Auth.Logout)

    return router
}

