package router

import (
	"net/http"
	"sentinel/packages/infrastructure/config"
	Auth "sentinel/packages/presentation/api/http/controllers/auth"
	User "sentinel/packages/presentation/api/http/controllers/user"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// i could just explicitly pass empty string in routes when i need it
// but it looks really awful, shitty and not obvious
const groupRootPath = ""

func Create() *echo.Echo {
	router := echo.New()

    router.HideBanner = true
    router.HidePort = true

    router.HTTPErrorHandler = handleHttpError
    router.JSONSerializer = serializer{}
    router.Binder = &binder{}

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

    userGroup := router.Group("/user")

    userGroup.POST(groupRootPath, User.Create)
    userGroup.DELETE(groupRootPath, User.SoftDelete)
    userGroup.POST("/restore", User.Restore)
    userGroup.DELETE("/drop", User.Drop)

    userGroup.PATCH("/login", User.ChangeLogin)
    userGroup.PATCH("/password", User.ChangePassword)
    userGroup.PATCH("/roles", User.ChangeRoles)

    return router
}

