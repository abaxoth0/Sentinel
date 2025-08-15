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
	"sentinel/packages/presentation/api/http/middleware"
	"sentinel/packages/presentation/api/http/request"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
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

	cors := echoMiddleware.CORSConfig{
		Skipper:      echoMiddleware.DefaultSkipper,
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

	router.Use(middleware.SecurityHeaders)
	router.Use(echoMiddleware.HTTPSRedirect())
	router.Use(echoMiddleware.BodyLimit("1M"))
	router.Use(echoMiddleware.CORSWithConfig(cors))
	router.Use(echoMiddleware.RequestID())
	router.Use(request.Middleware)
	router.Use(middleware.CheckOrigin)
	router.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))

	if config.Debug.Enabled {
		router.Use(echoMiddleware.Logger())
	}

	limit := middleware.NewRateLimiter()

	apiV1 := router.Group("/v1")

	// Path is strange, but it's convention from OpenID Connect Discovery (OIDC)
	apiV1.GET(
		"/.well-known/jwks.json", Auth.GetJWKs,
		limit.Max10reqPerSecond(middleware.InsignificantEndpoint),
	)

	authGroup := apiV1.Group("/auth", middleware.NoCache)

	authGroup.GET("/csrf-token", Auth.GetCSRFToken)
	authGroup.GET(
		rootPath, Auth.Verify,
		limit.Max1reqPerSecond(middleware.DefaultEndpoint),
		middleware.Secure, middleware.CheckUserSync,
	)
	authGroup.POST(
		rootPath, Auth.Login,
		limit.Max5reqPerMinute(middleware.SensitiveEndpoint),
		middleware.DoubleSubmitCSRF,
	)
	authGroup.PUT(
		rootPath, Auth.Refresh,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.DoubleSubmitCSRF,
	)
	authGroup.DELETE(
		rootPath, Auth.Logout,
		limit.Max1reqPerSecond(middleware.DefaultEndpoint),
		middleware.DoubleSubmitCSRF,
	)
	authGroup.DELETE(
		"/:sessionID", Auth.Logout,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	authGroup.DELETE(
		"/sessions/:uid", Auth.RevokeAllUserSessions,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	authGroup.POST(
		"/forgot-password", Auth.ForgotPassword,
		limit.Max5reqPerHour(middleware.SensitiveEndpoint),
	)
	authGroup.POST(
		"/reset-password", Auth.ResetPassword,
		limit.Max3reqPerMinute(middleware.SensitiveEndpoint),
		middleware.DoubleSubmitCSRF,
	)

	oauthSubGroup := authGroup.Group("/oauth", middleware.NoCache)

	oauthSubGroup.POST(
		"/introspect", OAuth.IntrospectOAuthToken,
		limit.Max1reqPerSecond(middleware.DefaultEndpoint),
		middleware.Secure, middleware.CheckUserSync,
	)
	oauthSubGroup.GET(
		"/google/login", OAuth.GoogleLogin,
		limit.Max5reqPerMinute(middleware.SensitiveEndpoint),
	)
	oauthSubGroup.GET(
		"/google/callback", OAuth.GoogleCallback,
		limit.Max5reqPerMinute(middleware.SensitiveEndpoint),
	)

	userGroup := apiV1.Group("/user", middleware.NoCache)

	userGroup.POST(
		rootPath, User.Create,
		limit.Max3reqPerMinute(middleware.DefaultEndpoint),
	)
	userGroup.DELETE(
		"/:uid", User.SoftDelete,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.PUT(
		"/:uid/restore", User.Restore,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.DELETE(
		rootPath, User.BulkSoftDelete,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.PUT(
		rootPath, User.BulkRestore,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.DELETE(
		"/:uid/drop", User.Drop,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.DELETE(
		"/all/drop", User.DropAllDeleted,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.POST(
		"/login/available", User.IsLoginAvailable,
		limit.Max1reqPerSecond(middleware.InsignificantEndpoint),
	)
	userGroup.GET(
		"/:uid/roles", User.GetRoles,
		limit.Max1reqPerSecond(middleware.DefaultEndpoint),
		middleware.Secure, middleware.CheckUserSync,
	)
	userGroup.PATCH(
		"/:uid/login", User.ChangeLogin,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.PATCH(
		"/:uid/password", User.ChangePassword,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.PATCH(
		"/:uid/roles", User.ChangeRoles,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync, middleware.DoubleSubmitCSRF,
	)
	userGroup.GET(
		"/activation/:token", Activation.Activate,
		limit.Max5reqPerMinute(middleware.DefaultEndpoint),
	)
	userGroup.PUT(
		"/activation/resend", Activation.Resend,
		limit.Max1reqPer5Minutes(middleware.DefaultEndpoint),
	)
	userGroup.GET(
		"/search", User.SearchUsers,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync,
	)
	userGroup.GET(
		"/:uid/sessions", User.GetUserSessions,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync,
	)
	userGroup.GET(
		"/:uid", User.GetUser,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.Secure, middleware.CheckUserSync,
	)

	rolesGroup := apiV1.Group("/roles", middleware.Secure, middleware.CheckUserSync)

	rolesGroup.GET(
		"/:serviceID", Roles.GetAll,
		limit.Max1reqPerSecond(middleware.InsignificantEndpoint),
	)

	cacheGroup := apiV1.Group("/cache", middleware.Secure, middleware.CheckUserSync, middleware.NoCache)

	cacheGroup.DELETE(
		rootPath, Cache.Drop,
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.DoubleSubmitCSRF,
	)

	docsGroupMiddlewares := []echo.MiddlewareFunc{
		limit.Max1reqPerSecond(middleware.SensitiveEndpoint),
		middleware.DoubleSubmitCSRF,
	}
	if !config.Debug.Enabled {
		docsGroupMiddlewares = append(docsGroupMiddlewares, middleware.Secure, middleware.CheckUserSync)
	}

	docsGroup := router.Group("/docs", docsGroupMiddlewares...)

	docsGroup.GET("/*", Docs.Swagger)

	return router
}

