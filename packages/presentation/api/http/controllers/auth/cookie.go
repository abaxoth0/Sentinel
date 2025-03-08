package authcontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/config"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/response"
	"time"

	"github.com/labstack/echo/v4"
)

func deleteCookie(ctx echo.Context, cookie *http.Cookie) {
    cookie.Expires = time.Now().Add(time.Hour * -1)

    ctx.SetCookie(cookie)
}

func newAuthCookie(refreshToken *token.SignedToken) *http.Cookie {
    return &http.Cookie{
		Name:     token.RefreshTokenKey,
		Value:    refreshToken.Value,
		Path:     "/",
        // token's TTL specified in millisconds,
        // but MaxAge expects time in seconds
		MaxAge:   int(refreshToken.TTL) / 1000,
		HttpOnly: true,
		Secure:   config.HTTP.Secured,
	}
}

func getAuthCookie(ctx echo.Context) (*http.Cookie, error) {
    authCookie, err := ctx.Cookie(token.RefreshTokenKey)

    if err != nil {
        if err == http.ErrNoCookie {
            return nil, response.Unauthorized
        }

        return nil, err
    }

    return authCookie, nil
}

