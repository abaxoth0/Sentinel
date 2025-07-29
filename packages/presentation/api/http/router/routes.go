package router

import (
	"net/http"
	_ "sentinel/docs"
	"sentinel/packages/common/config"
	"sentinel/packages/common/logger"
	Activation "sentinel/packages/presentation/api/http/controllers/activation"
	Auth "sentinel/packages/presentation/api/http/controllers/auth"
	Cache "sentinel/packages/presentation/api/http/controllers/cache"
	Docs "sentinel/packages/presentation/api/http/controllers/docs"
	Roles "sentinel/packages/presentation/api/http/controllers/roles"
	User "sentinel/packages/presentation/api/http/controllers/user"
	"sentinel/packages/presentation/api/http/request"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var routerLogger = logger.NewSource("ROUTER", logger.Default)

// i could just explicitly pass empty string in routes when i need it
// but it looks really awful, shitty and not obvious
const rootPath = ""

func Create() *echo.Echo {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.Secret.SentryDSN,
		EnableTracing: true,
		TracesSampleRate: config.Sentry.TraceSampleRate,
		Debug: config.Debug.Enabled,
		ServerName: config.App.ServiceID,
		AttachStacktrace: true,
	}); err != nil {
		panic("Sentry initialization failed: " + err.Error())
	}

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
		AllowHeaders: []string{
			"X-CSRF-Token",
		},
    }

	router.Use(securityHeaders)
	router.Use(middleware.BodyLimit("1M"))
	router.Use(middleware.HTTPSRedirect())
    router.Use(middleware.CORSWithConfig(cors))
	router.Use(middleware.RequestID())
	router.Use(request.Middleware)
	router.Use(checkOrigin)
	router.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))
    // router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10_000)))

    if config.Debug.Enabled {
        router.Use(middleware.Logger())
    }

	// Path is strange, but it's convention from OpenID Connect Discovery (OIDC)
	router.GET("/.well-known/jwks.json", Auth.GetJWKs)

    authGroup := router.Group("/auth", noCache)

	authGroup.GET("/csrf-token", Auth.GetCSRFToken)
    authGroup.GET(rootPath, Auth.Verify, secure, preventUserDesync)
    authGroup.POST(rootPath, Auth.Login, doubleSubmitCSRF)
    authGroup.PUT(rootPath, Auth.Refresh, doubleSubmitCSRF)
    authGroup.DELETE(rootPath, Auth.Logout, doubleSubmitCSRF)
	authGroup.DELETE("/:sessionID", Auth.Logout, secure, preventUserDesync, doubleSubmitCSRF)
	authGroup.DELETE("/sessions/:uid", Auth.RevokeAllUserSessions, secure, preventUserDesync, doubleSubmitCSRF)
	authGroup.POST("/oauth/introspect", Auth.IntrospectOAuthToken, secure, preventUserDesync)

    userGroup := router.Group("/user", secure, preventUserDesync, noCache)

    userGroup.POST(rootPath, User.Create)
    userGroup.DELETE("/:uid", User.SoftDelete, doubleSubmitCSRF)
    userGroup.PUT("/:uid/restore", User.Restore, doubleSubmitCSRF)
    userGroup.DELETE(rootPath, User.BulkSoftDelete, doubleSubmitCSRF)
    userGroup.PUT(rootPath, User.BulkRestore, doubleSubmitCSRF)
    userGroup.DELETE("/:uid/drop", User.Drop, doubleSubmitCSRF)
    userGroup.DELETE("/all/drop", User.DropAllDeleted, doubleSubmitCSRF)
    userGroup.POST("/login/available", User.IsLoginAvailable)
    userGroup.GET("/:uid/roles", User.GetRoles)
    userGroup.PATCH("/:uid/login", User.ChangeLogin, doubleSubmitCSRF)
    userGroup.PATCH("/:uid/password", User.ChangePassword, doubleSubmitCSRF)
    userGroup.PATCH("/:uid/roles", User.ChangeRoles, doubleSubmitCSRF)
    userGroup.GET("/activation/:token", Activation.Activate)
    userGroup.PUT("/activation/resend", Activation.Resend)
	userGroup.GET("/search", User.SearchUsers)
	userGroup.GET("/:uid/sessions", User.GetUserSessions)
	userGroup.GET("/:uid", User.GetUser)

    rolesGroup := router.Group("/roles", secure, preventUserDesync)

    rolesGroup.GET("/:serviceID", Roles.GetAll)

    cacheGroup := router.Group("/cache", secure, preventUserDesync, noCache)

    cacheGroup.DELETE(rootPath, Cache.Drop, doubleSubmitCSRF)

	docsGroupMiddlewares := []echo.MiddlewareFunc{doubleSubmitCSRF}
	if !config.Debug.Enabled {
		docsGroupMiddlewares = append(docsGroupMiddlewares, secure, preventUserDesync)
	}

	docsGroup := router.Group("/docs", docsGroupMiddlewares...)

	docsGroup.GET("/*", Docs.Swagger)

    return router
}

