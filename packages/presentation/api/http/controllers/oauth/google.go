package oauthcontroller

import (
	"context"
	"net/http"
	"sentinel/packages/common/config"
	"sentinel/packages/common/encoding/json"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/presentation/api"
	controller "sentinel/packages/presentation/api/http/controllers"
	SharedController "sentinel/packages/presentation/api/http/controllers/shared"
	"sentinel/packages/presentation/api/http/request"
	"strconv"
	"time"

	"github.com/abaxoth0/go-pwgen"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config

func initGoogle() {
	googleOAuthConfig = &oauth2.Config{
		RedirectURL: api.GetBaseURL() + "/v1/auth/oauth/google/callback",
		ClientID: config.Secret.OAuthGoogleClientID,
		ClientSecret: config.Secret.OAuthGoogleClientSecret,
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}

// @Summary 		Login/Signup into the service using google account
// @Description 	Alternative login endpoint, redirects to the google OAuth2.0 endpoint.
// @ID 				google-login
// @Tags			third-party-auth,oauth
// @Accept			json
// @Produce			json
// @Success			307
// @Failure			500 	{object} 	responsebody.Error
// @Router			/v1/auth/oauth/google/login [get]
func GoogleLogin(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	if !isInit {
		controller.Log.Panic("Failed to handle google login", "OAuth controller wasn't initialized", reqMeta)
		return Error.StatusInternalError
	}

	state, err := SharedController.NewCSRFToken(ctx)
	if err != nil {
		return err
	}

	sessionID := uuid.New().String()
	session := newOAuthSession(ctx.RealIP(), state, ctx.Request().UserAgent())
	if err := sessionStore.Save(googleProvider, sessionID, &session); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	ctx.SetCookie(&http.Cookie{
		Name: "oauth_session",
		Value: sessionID,
		HttpOnly: true,
		Secure: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := googleOAuthConfig.AuthCodeURL(state)

	return ctx.Redirect(http.StatusTemporaryRedirect, url)
}
/*
	Response from https://www.googleapis.com/oauth2/v3/userinfo

	Contains not all fields from response, only those which are needed.
	Response example:
	{
		"sub": "google_account_id",
		"name": "fullname",
		"given_name": "firstname",
		"family_name": "lastname",
		"picture": "url",
		"email": "admin@email.com",
		"email_verified": true
	}
*/
type googleUserInfoPartialResponse struct {
	GoogleID 		string 	`json:"sub"`
	Email 	 		string	`json:"email"`
	IsEmailVerified bool 	`json:"email_verified"`
}

// @Summary 		Actual handler for auth via google account
// @Description 	Do not access this endpoint manually, google API will automatically redirect to it. E-Mail of google account MUST be verified.
// @ID 				google-login-handler
// @Tags			third-party-auth,oauth
// @Param 			code query string true "Short-lived, temporary authorization code issued by Google's authorization server"
// @Param 			state query string true "OAuth state token"
// @Accept			json
// @Produce			json
// @Success			200 			{object} 	responsebody.Token
// @Failure			400,403,408,500	{object} 	responsebody.Error
// @Router			/v1/auth/oauth/google/callback [get]
// @Security 		OAuthSession
func GoogleCallback(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	if !isInit {
		controller.Log.Panic("Failed to handle google login", "OAuth controller wasn't initialized", reqMeta)
		return Error.StatusInternalError
	}

	code := ctx.QueryParam("code")
	if code == "" {
		errMsg := "Missing query param: code"
		controller.Log.Error("Login failed", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMsg)
	}

	if err := validateOAuthSession(ctx, googleProvider); err != nil {
		controller.Log.Error("Login failed: suspicious context change", err.Error(), reqMeta)
		return echo.NewHTTPError(http.StatusForbidden, "Security vioalation detected")
	}

	c, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()

	oauthToken, err := googleOAuthConfig.Exchange(c, code)
	if err != nil {
		controller.Log.Error("Login failed", err.Error(), reqMeta)
		return Error.StatusInternalError
	}

	client := oauth2.NewClient(c, googleOAuthConfig.TokenSource(c, oauthToken))

	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		controller.Log.Error("Failed to get user info from google API", err.Error(), reqMeta)
		return Error.StatusInternalError
	}
	defer userInfo.Body.Close()

	body, err := json.Decode[googleUserInfoPartialResponse](userInfo.Body)
	if err != nil {
		return Error.StatusInternalError
	}

	if !body.IsEmailVerified {
		errMsg := "Your email isn't verified by google"
		controller.Log.Error("Login failed", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusForbidden, errMsg)
	}

	user, e := DB.Database.GetUserByLogin(body.Email)
	if e != nil && e != Error.StatusNotFound {
		return e
	}

	// User with this email doesn't exists -> create new user
	if e == Error.StatusNotFound {
		controller.Log.Trace("Generating password for new user...", reqMeta)

		password, err := pwgen.Generate(32, pwgen.LOWER|pwgen.UPPER|pwgen.DIGITS)
		if err != nil {
			errMsg := "Failed to generate password"
			controller.Log.Error("Failed to generate password for new user", errMsg, reqMeta)
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}

		controller.Log.Trace("Generating password for new user: OK", reqMeta)

		uid, e := DB.Database.Create(body.Email, password)
		if e != nil {
			return e
		}

		// There may occur a problem since DB.Database.Create() creates user in primary DB,
		// but DB.Database.FindUserByID() tries to find user in replica DB.
		// This is fine, but there are very small time gap between this two actions,
		// so this loop is needed to give replica DB enough time to synchronize with primary DB
		retries := 0
		maxRetries := 100
		retryWaitTime := time.Millisecond * 50

		// Max wait time = maxRetries * retryWaitTime
		for retries <= maxRetries {
			user, e = DB.Database.GetUserByID(uid)
			if e != nil && e != Error.StatusNotFound {
				return e
			}
			if e == nil {
				break
			}
			// err == Error.StatusNotFound
			time.Sleep(retryWaitTime)
			retries++
		}

		controller.Log.Trace("Retries of searching for a created user: " + strconv.Itoa(retries), reqMeta)

		if retries == maxRetries {
			controller.Log.Error(
				"Failed to find created user",
				"Timeout: replica DB didn't have enough time to synchronize with primary DB",
				reqMeta,
			)
			return echo.NewHTTPError(
				Error.StatusTimeout.Status(),
				"Account has been created, but login failed.",
			)
		}

		return SharedController.AuthenticateWithNewSession(ctx, user, []string{config.Auth.SelfAudience})
	}

	return SharedController.Authenticate(ctx, user, []string{config.Auth.SelfAudience})
}
