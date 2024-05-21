package net

import (
	"net/http"
	"sentinel/packages/config"
	"sentinel/packages/models/token"
)

type cookie struct{}

var Cookie = cookie{}

func (c cookie) BuildAuth(refreshToken token.SignedToken) *http.Cookie {
	return &http.Cookie{
		Name:     token.RefreshTokenKey,
		Value:    refreshToken.Value,
		Path:     "/",
		MaxAge:   int(refreshToken.TTL),
		HttpOnly: true,
		Secure:   config.Debug.Enabled,
	}
}

func (c cookie) Delete(cookie *http.Cookie, w http.ResponseWriter) {
	cookie.HttpOnly = true
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}
