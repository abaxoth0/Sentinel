package controller

import (
	"fmt"
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/util"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/request"
	RequestBody "sentinel/packages/presentation/data/request"
	"github.com/labstack/echo/v4"
	"sentinel/packages/common/config"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"

	"github.com/golang-jwt/jwt/v5"
)

var Logger = logger.NewSource("CONTROLLER", logger.Default)

func BindAndValidate[T RequestBody.Validator](ctx echo.Context, dest T) error {
    reqMeta := request.GetMetadata(ctx)

    Logger.Trace("Binding and validating request...", reqMeta)

    if err := ctx.Bind(&dest); err != nil {
        Logger.Error("Failed to bind request", err.Error(), reqMeta)
        return err
    }

    if err := dest.Validate(); err != nil {
        Logger.Error("Request validation failed", err.Error(), reqMeta)
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    Logger.Trace("Binding and validating request: OK", reqMeta)

    return nil
}

type wwwAuthenticateParamas struct {
    Realm string
    Error string
    ErrorDescription string
}

func applyWWWAuthenticate(ctx echo.Context, params *wwwAuthenticateParamas) {
    ctx.Response().Header().Set(
        "WWW-Authenticate",
        `Bearer realm="`+params.Realm+`", error="`+params.Error+`", error_description="`+params.ErrorDescription+`"`,
    )
}

func HandleTokenError(ctx echo.Context, err *Error.Status) *echo.HTTPError {
	reqMeta := request.GetMetadata(ctx)

	Logger.Trace("Handling token error...", reqMeta)

    // token persist, but invalid
    if token.IsTokenError(err) {
        applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
            Realm: "api",
            Error: util.Ternary(err == token.TokenExpired, "expired_token", "invalid_token"),
            ErrorDescription: err.Error(),
        })

        authCookie, err := GetAuthCookie(ctx)
        if err != nil {
			Logger.Trace("Handling token error: OK", reqMeta)
            return err
        }

        DeleteCookie(ctx, authCookie)
        // token is missing
    } else if err == Error.StatusUnauthorized {
        applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
            Realm: "api",
            Error: "invalid_request",
            ErrorDescription: "No token provided",
        })
    }

	Logger.Trace("Handling token error: OK", reqMeta)

    return ConvertErrorStatusToHTTP(err)
}

var invalidAuthorizationHeaderFormat = Error.NewStatusError(
    "Invalid Authorization header. Expected format: 'Bearer <token>'",
    http.StatusUnauthorized,
)

// IMPORTANT: This function can only be used if the route has been secured (via the 'secured' middleware).
// Otherwise, using this function will cause panic.
func getNonNilValueFromSecuredRequestContext[T any](ctx echo.Context, key string) T {
	value := ctx.Get(key)
	if value == nil {
		secured := ctx.Get("Secured")
		switch s := secured.(type) {
		case bool:
			Logger.Panic(
				"Failed to get value from request context",
				fmt.Sprintf("Route %s %s isn't secured", ctx.Request().Method, ctx.Path()),
				nil,
			)
		default:
			Logger.Panic(
				"Failed to get value from request context",
				fmt.Sprintf("Secured has invalid type: Expected bool, but got %T", s),
				nil,
			)
		}
		Logger.Panic(
			"Failed to get value from request context",
			"value is nil",
			nil,
		)
		return *new(T) // Anyway will panic
	}
	switch v := value.(type) {
	case T:
		return v
	default:
		Logger.Panic(
			"Failed to get value from request context",
			fmt.Sprintf("value has invalid type: %T", v),
			nil,
		)
		return *new(T) // Anyway will panic
	}
}

// IMPORTANT: This function can only be used if the route has been secured (via the 'secured' middleware).
// Otherwise, using this function will cause panic.
//
// Returned value guaranteed to be non-nil.
func GetAccessToken(ctx echo.Context) *jwt.Token {
	return getNonNilValueFromSecuredRequestContext[*jwt.Token](ctx, "access_token")
}

// IMPORTANT: This function can only be used if the route has been secured (via the 'secured' middleware).
// Otherwise, using this function will cause panic.
//
// Returned value guaranteed to be non-nil.
func GetAccessTokenPayload(ctx echo.Context) *UserDTO.Payload {
	return getNonNilValueFromSecuredRequestContext[*UserDTO.Payload](ctx, "access_token_payload")
}

// IMPORTANT: This function can only be used if the route has been secured (via the 'secured' middleware).
// Otherwise, using this function will cause panic.
//
// Returned value guaranteed to be non-nil.
func GetBasicAction(ctx echo.Context) *ActionDTO.Basic {
	return getNonNilValueFromSecuredRequestContext[*ActionDTO.Basic](ctx, "basic_action")
}

const RefreshTokenCookieKey string = "refreshToken"

// Retrieves and validates refresh token.
//
// Returns pointer to token and nil if valid and not expired token was found.
// Otherwise returns empty pointer to token and *Error.Status.
func GetRefreshToken(ctx echo.Context) (*jwt.Token, *Error.Status) {
	reqMeta := request.GetMetadata(ctx)

	Logger.Trace("Getting refresh token from the request...", reqMeta)

    cookie, err := ctx.Cookie(RefreshTokenCookieKey)
    if err != nil {
        if err == http.ErrNoCookie {
            return nil, Error.StatusUnauthorized
        }

        Logger.Error("Failed to get auth cookie", err.Error(), reqMeta)
        return nil, Error.StatusInternalError
    }

	token, e := token.ParseSingedToken(cookie.Value, config.Secret.RefreshTokenPublicKey)
    if e != nil {
        return nil, e
    }

	Logger.Trace("Getting refresh token from the request: OK", reqMeta)

	return token, nil
}

