package controller

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/request"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var invalidAuthorizationHeaderFormat = Error.NewStatusError(
    "Invalid Authorization header. Expected format: 'Bearer <token>'",
    http.StatusUnauthorized,
)

// Retrieves and validates access token.
//
// Returns token pointer and nil if valid and not expired token was found.
// Otherwise returns empty token pointer and error.
func GetAccessToken(ctx echo.Context) (*jwt.Token, *Error.Status) {
	reqMeta, e := request.GetLogMeta(ctx)
	if e != nil {
		Logger.Panic("Failed to get log meta for the request", e.Error(), nil)
		return nil, Error.NewStatusError(e.Error(), http.StatusInternalServerError)
	}

	Logger.Trace("Getting access token from the request...", reqMeta)

    authHeader := ctx.Request().Header.Get("Authorization")
	if strings.ReplaceAll(authHeader, " ", "") == "" {
		return nil, Error.StatusUnauthorized
	}
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return nil, invalidAuthorizationHeaderFormat
    }

    splitAuthHeader := strings.Split(authHeader, "Bearer ")
    if len(splitAuthHeader) != 2 {
        return nil, invalidAuthorizationHeaderFormat
    }

    accessTokenStr := splitAuthHeader[1]

	token, err := token.ParseSingedToken(accessTokenStr, config.Secret.AccessTokenPublicKey)
    if err != nil {
        return nil, err
    }

	Logger.Trace("Getting access token from the request: OK", reqMeta)

	return token, nil
}

const RefreshTokenCookieKey string = "refreshToken"

// Retrieves and validates refresh token.
//
// Returns pointer to token and nil if valid and not expired token was found.
// Otherwise returns empty pointer to token and *Error.Status.
func GetRefreshToken(ctx echo.Context) (*jwt.Token, *Error.Status) {
	reqMeta, err := request.GetLogMeta(ctx)
	if err != nil {
		Logger.Panic("Failed to get log meta for the request", err.Error(), nil)
		return nil, Error.NewStatusError(err.Error(), http.StatusInternalServerError)
	}

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

