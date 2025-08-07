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
	OAuth "sentinel/packages/presentation/api/http/controllers/oauth"
	Roles "sentinel/packages/presentation/api/http/controllers/roles"
	User "sentinel/packages/presentation/api/http/controllers/user"
	"sentinel/packages/presentation/api/http/request"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var log = logger.NewSource("ROUTER", logger.Default)

// i could just explicitly pass empty string in routes when i need it
// but it looks really awful, shitty and not obvious
const rootPath = ""

func Create() *echo.Echo {
	OAuth.Init()

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

	apiV1 := router.Group("/v1")

	// Path is strange, but it's convention from OpenID Connect Discovery (OIDC)
	apiV1.GET("/.well-known/jwks.json", Auth.GetJWKs)

    authGroup := apiV1.Group("/auth", noCache)

	authGroup.GET("/csrf-token", Auth.GetCSRFToken)
    authGroup.GET(rootPath, Auth.Verify, secure, preventUserDesync)
    authGroup.POST(rootPath, Auth.Login, doubleSubmitCSRF)
    authGroup.PUT(rootPath, Auth.Refresh, doubleSubmitCSRF)
    authGroup.DELETE(rootPath, Auth.Logout, doubleSubmitCSRF)
	authGroup.DELETE("/:sessionID", Auth.Logout, secure, preventUserDesync, doubleSubmitCSRF)
	authGroup.DELETE("/sessions/:uid", Auth.RevokeAllUserSessions, secure, preventUserDesync, doubleSubmitCSRF)

	oauthSubGroup := authGroup.Group("/oauth", noCache)

	oauthSubGroup.POST("/introspect", OAuth.IntrospectOAuthToken, secure, preventUserDesync)
	oauthSubGroup.GET("/google/login", OAuth.GoogleLogin)
	oauthSubGroup.GET("/google/callback", OAuth.GoogleCallback)

    userGroup := apiV1.Group("/user", noCache)

    userGroup.POST(rootPath, User.Create)
    userGroup.DELETE("/:uid", User.SoftDelete, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.PUT("/:uid/restore", User.Restore, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.DELETE(rootPath, User.BulkSoftDelete, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.PUT(rootPath, User.BulkRestore, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.DELETE("/:uid/drop", User.Drop, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.DELETE("/all/drop", User.DropAllDeleted, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.POST("/login/available", User.IsLoginAvailable)
    userGroup.GET("/:uid/roles", User.GetRoles, secure, preventUserDesync)
    userGroup.PATCH("/:uid/login", User.ChangeLogin, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.PATCH("/:uid/password", User.ChangePassword, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.PATCH("/:uid/roles", User.ChangeRoles, secure, preventUserDesync, doubleSubmitCSRF)
    userGroup.GET("/activation/:token", Activation.Activate, secure, preventUserDesync)
    userGroup.PUT("/activation/resend", Activation.Resend, secure, preventUserDesync)
	userGroup.GET("/search", User.SearchUsers, secure, preventUserDesync)
	userGroup.GET("/:uid/sessions", User.GetUserSessions, secure, preventUserDesync)
	userGroup.GET("/:uid", User.GetUser, secure, preventUserDesync)

    rolesGroup := apiV1.Group("/roles", secure, preventUserDesync)

    rolesGroup.GET("/:serviceID", Roles.GetAll)

    cacheGroup := apiV1.Group("/cache", secure, preventUserDesync, noCache)

    cacheGroup.DELETE(rootPath, Cache.Drop, doubleSubmitCSRF)

	docsGroupMiddlewares := []echo.MiddlewareFunc{doubleSubmitCSRF}
	if !config.Debug.Enabled {
		docsGroupMiddlewares = append(docsGroupMiddlewares, secure, preventUserDesync)
	}

	docsGroup := router.Group("/docs", docsGroupMiddlewares...)

	docsGroup.GET("/*", Docs.Swagger)

    return router
}

