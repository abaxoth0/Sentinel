package controller

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/request"
	"time"

	"github.com/labstack/echo/v4"
)

func DeleteCookie(ctx echo.Context, cookie *http.Cookie) {
    cookie.Expires = time.Now().Add(time.Hour * -1)

    ctx.SetCookie(cookie)
}

func GetAuthCookie(ctx echo.Context) (*http.Cookie, *echo.HTTPError) {
	reqMeta := request.GetMetadata(ctx)

	Logger.Trace("Getting auth cookie...", reqMeta)

    authCookie, err := ctx.Cookie(refreshTokenCookieKey)

    if err != nil {
		Logger.Error("Failed to get auth cookie", err.Error(), reqMeta)

        if err == http.ErrNoCookie {
            return nil, ConvertErrorStatusToHTTP(Error.StatusUnauthorized)
        }
        return nil, ConvertErrorStatusToHTTP(Error.StatusInternalError)
    }

	Logger.Trace("Getting auth cookie: OK", reqMeta)

    return authCookie, nil
}

func NewAuthCookie(refreshToken *token.SignedToken) *http.Cookie {
    return &http.Cookie{
		Name:     refreshTokenCookieKey,
		Value:    refreshToken.String(),
		Path:     "/",
        // token's TTL specified in millisconds,
        // but MaxAge expects time in seconds
		MaxAge:   int(refreshToken.TTL()) / 1000,
		HttpOnly: true,
		Secure:   config.HTTP.Secured,
	}
}

