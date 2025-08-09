package oauthcontroller

import (
	"errors"
	"sentinel/packages/infrastructure/cache"
	controller "sentinel/packages/presentation/api/http/controllers"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Used for security.
// Allows to ensure that IP, User Agent and State (CSRF token) match during redirects.
type oauthSession struct {
	IP 			string
	// CSRF token (Name comes from OAuth convention. See RFC 6749 p4.1.1)
	State 		string
	UserAgent	string
}

func newOAuthSession(ip, state, userAgent string) oauthSession {
	return oauthSession{
		IP: ip,
		State: state,
		UserAgent: userAgent,
	}
}

func validateOAuthSession(ctx echo.Context, provider authProvider) error {
	sessionCookie, err := ctx.Cookie("oauth_session")
	if err != nil {
		return errors.New("OAuth session cookie is missing")
	}

	oauthSession, ok := sessionStore.Extract(provider, sessionCookie.Value)
	if !ok {
		return errors.New("OAuth session wasn't found")
	}
	if oauthSession.IP != ctx.RealIP() {
		return errors.New("OAuth session IP mismatch")
	}
	if ctx.QueryParam("state") != oauthSession.State {
		return errors.New("Invalid state token")
	}
	if ctx.Request().UserAgent() != oauthSession.UserAgent {
		return errors.New("User agent mismatch")
	}

	return nil
}

type oauthSessionStore struct {
	//
}

var oauthSessionTTL = time.Minute * 5

func (s *oauthSessionStore) Save(provider authProvider, id string, session *oauthSession) error {
	id = provider.String()+"_"+id
	value := session.IP+"\n"+session.State+"\n"+session.UserAgent

	if err := cache.Client.SetWithTTL(id, value, oauthSessionTTL); err != nil {
		errMsg := "Failed to store oauth session"
		controller.Log.Error(errMsg, err.Error(), nil)
		return errors.New(errMsg)
	}

	return nil
}

// IMPORTANT: This method will delete session from store if specified id was found
func (s *oauthSessionStore) Extract(provider authProvider, id string) (oauthSession, bool) {
	key := provider.String()+"_"+id

	sessionStr, hit := cache.Client.Get(key)
	if !hit {
		var zero oauthSession
		return zero, false
	} else {
		cache.Client.Delete(key)
	}

	sessionData := strings.Split(sessionStr, "\n")

	if len(sessionData) != 3 {
		var zero oauthSession
		return zero, false
	}

	return oauthSession{
		IP: sessionData[0],
		State: sessionData[1],
		UserAgent: sessionData[2],
	}, true
}

var sessionStore = new(oauthSessionStore)

