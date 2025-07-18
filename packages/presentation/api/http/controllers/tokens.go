package controller

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	actiondto "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	usermapper "sentinel/packages/infrastructure/mappers/user"
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
	reqMeta := request.GetMetadata(ctx)

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

	payload, err := usermapper.PayloadFromClaims(token.Claims.(jwt.MapClaims))
	if err != nil {
		return nil, err
	}

	act := actiondto.NewUserTargeted(payload.ID, payload.ID, payload.Roles)

	if _, err := DB.Database.GetSessionByID(act, payload.SessionID, true); err == nil {
		return nil, Error.StatusSessionRevoked
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

