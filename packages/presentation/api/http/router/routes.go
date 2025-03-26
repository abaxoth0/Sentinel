package router

import (
	"net/http"
	"sentinel/packages/config"
	Auth "sentinel/packages/presentation/api/http/controllers/auth"
	Cache "sentinel/packages/presentation/api/http/controllers/cache"
	Roles "sentinel/packages/presentation/api/http/controllers/roles"
	User "sentinel/packages/presentation/api/http/controllers/user"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// i could just explicitly pass empty string in routes when i need it
// but it looks really awful, shitty and not obvious
const rootPath = ""

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

    authGroup.GET(rootPath, Auth.Verify)
    authGroup.POST(rootPath, Auth.Login)
    authGroup.PUT(rootPath, Auth.Refresh)
    authGroup.DELETE(rootPath, Auth.Logout)

    userGroup := router.Group("/user")

    userGroup.POST(rootPath, User.Create)
    userGroup.DELETE(rootPath, User.SoftDelete)
    userGroup.POST("/restore", User.Restore)
    userGroup.DELETE("/drop", User.Drop)
    userGroup.POST("/login/check", User.IsLoginExists)

    userGroup.PATCH("/login", User.ChangeLogin)
    userGroup.PATCH("/password", User.ChangePassword)
    userGroup.PATCH("/roles", User.ChangeRoles)
    userGroup.GET("/roles", User.GetRoles)

    rolesGroup := router.Group("/roles")

    rolesGroup.GET("/:serviceID", Roles.GetAll)

    cacheGroup := router.Group("/cache")

    cacheGroup.DELETE(rootPath, Cache.Drop)

    return router
}

