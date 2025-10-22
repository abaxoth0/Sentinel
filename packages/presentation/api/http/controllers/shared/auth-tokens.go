package sharedcontroller

import (
	"fmt"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/cookie"
	"sentinel/packages/presentation/api/http/request"

	"github.com/labstack/echo/v4"

	"github.com/golang-jwt/jwt/v5"
)

// IMPORTANT: This function can only be used if the route has been secured (via the 'secured' middleware).
// Otherwise this function will cause panic.
func getNonNilValueFromSecuredContext[T any](ctx echo.Context, key string) T {
	value := ctx.Get(key)
	if value == nil {
		secured := ctx.Get("Secured")
		switch s := secured.(type) {
		case bool:
			controller.Log.Panic(
				"Failed to get value from request context",
				fmt.Sprintf("Route %s %s isn't secured", ctx.Request().Method, ctx.Path()),
				nil,
			)
		default:
			controller.Log.Panic(
				"Failed to get value from request context",
				fmt.Sprintf("Secured has invalid type: Expected bool, but got %T", s),
				nil,
			)
		}
		controller.Log.Panic(
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
		controller.Log.Panic(
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
func GetUserPayload(ctx echo.Context) *UserDTO.Payload {
	return getNonNilValueFromSecuredContext[*UserDTO.Payload](ctx, "user_payload")
}

// IMPORTANT: This function can only be used if the route has been secured (via the 'secured' middleware).
// Otherwise, using this function will cause panic.
//
// Returned value guaranteed to be non-nil.
func GetBasicAction(ctx echo.Context) *ActionDTO.Basic {
	return getNonNilValueFromSecuredContext[*ActionDTO.Basic](ctx, "basic_action")
}

// Retrieves and validates refresh token.
//
// Returns pointer to token and nil if valid and not expired token was found.
// Otherwise returns empty pointer to token and *Error.Status.
func GetRefreshToken(ctx echo.Context) (*jwt.Token, *Error.Status) {
	reqMeta := request.GetMetadata(ctx)

	controller.Log.Trace("Getting refresh token from the request...", reqMeta)

	cookie, err := ctx.Cookie(cookie.RefreshTokenCookieKey)
	if err != nil {
		controller.Log.Error("Failed to get refresh token from cookie", err.Error(), reqMeta)
		if err == http.ErrNoCookie {
			return nil, Error.StatusUnauthorized
		}
		return nil, Error.StatusInternalError
	}

	token, e := token.ParseSingedToken(cookie.Value, config.Secret.RefreshTokenPublicKey)
	if e != nil {
		return nil, e
	}

	controller.Log.Trace("Getting refresh token from the request: OK", reqMeta)

	return token, nil
}
