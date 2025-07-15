package router

import (
	"net/http"
	_ "sentinel/docs"
	"sentinel/packages/common/config"
	Activation "sentinel/packages/presentation/api/http/controllers/activation"
	Auth "sentinel/packages/presentation/api/http/controllers/auth"
	Cache "sentinel/packages/presentation/api/http/controllers/cache"
	Docs "sentinel/packages/presentation/api/http/controllers/docs"
	Roles "sentinel/packages/presentation/api/http/controllers/roles"
	User "sentinel/packages/presentation/api/http/controllers/user"
	"sentinel/packages/presentation/api/http/request"

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

	router.Use(request.Middleware)
    router.Use(middleware.CORSWithConfig(cors))
    router.Use(catchError)
	router.Use(preventUserDesync)
    // router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10_000)))

    if config.Debug.Enabled {
        router.Use(middleware.Logger())
    }

    authGroup := router.Group("/auth")

    authGroup.GET(rootPath, Auth.Verify)
    authGroup.POST(rootPath, Auth.Login)
    authGroup.PUT(rootPath, Auth.Refresh)
    authGroup.DELETE(rootPath, Auth.Logout)
	authGroup.DELETE("/:sessionID", Auth.Logout)
	authGroup.DELETE("/sessions/:uid", Auth.RevokeAllUserSessions)

    userGroup := router.Group("/user")

    userGroup.POST(rootPath, User.Create)
    userGroup.DELETE("/:uid", User.SoftDelete)
    userGroup.PUT("/:uid/restore", User.Restore)
    userGroup.DELETE(rootPath, User.BulkSoftDelete)
    userGroup.PUT(rootPath, User.BulkRestore)
    userGroup.DELETE("/:uid/drop", User.Drop)
    userGroup.DELETE("/all/drop", User.DropAllDeleted)
    userGroup.POST("/login/available", User.IsLoginAvailable)
    userGroup.GET("/:uid/roles", User.GetRoles)
    userGroup.PATCH("/:uid/login", User.ChangeLogin)
    userGroup.PATCH("/:uid/password", User.ChangePassword)
    userGroup.PATCH("/:uid/roles", User.ChangeRoles)
    userGroup.GET("/activation/:token", Activation.Activate)
    userGroup.PUT("/activation/resend", Activation.Resend)
	userGroup.GET("/search", User.SearchUsers)
	userGroup.GET("/:uid/sessions", User.GetUserSessions)

    rolesGroup := router.Group("/roles")

    rolesGroup.GET("/:serviceID", Roles.GetAll)

    cacheGroup := router.Group("/cache")

    cacheGroup.DELETE(rootPath, Cache.Drop)

	docsGroup := router.Group("/docs")

	docsGroup.GET("/*", Docs.Swagger)

    return router
}

