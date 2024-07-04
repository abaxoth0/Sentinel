package authcontroller

import (
	"net/http"
	"sentinel/packages/config"
	"sentinel/packages/models/token"
)

func buildAuthCookie(refreshToken *token.SignedToken) *http.Cookie {
	return &http.Cookie{
		Name:     token.RefreshTokenKey,
		Value:    refreshToken.Value,
		Path:     "/",
		MaxAge:   int(refreshToken.TTL),
		HttpOnly: true,
		Secure:   config.Debug.Enabled,
	}
}
