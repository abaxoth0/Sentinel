package authcontroller

import (
	"net/http"
	"sentinel/packages/common/config"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
)

func newAuthCookie(refreshToken *token.SignedToken) *http.Cookie {
    return &http.Cookie{
		Name:     controller.RefreshTokenCookieKey,
		Value:    refreshToken.String(),
		Path:     "/",
        // token's TTL specified in millisconds,
        // but MaxAge expects time in seconds
		MaxAge:   int(refreshToken.TTL()) / 1000,
		HttpOnly: true,
		Secure:   config.HTTP.Secured,
	}
}

