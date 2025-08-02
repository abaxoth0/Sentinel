package oauthcontroller

import (
	"context"
	"net/http"
	"sentinel/packages/common/config"
	"sentinel/packages/common/encoding/json"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	ResponseBody "sentinel/packages/presentation/data/response"
	"strconv"
	"time"

	"github.com/abaxoth0/go-pwgen"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config

var passwordGenerationOptions = &pwgen.Config{
	Length: 32,
	Digits: true,
	Lower: true,
	Upper: true,
}

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
		return controller.ConvertErrorStatusToHTTP(Error.StatusInternalError)
	}

	state, err := controller.NewCSRFToken(ctx)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	sessionID := uuid.New().String()
	session := newOAuthSession(ctx.RealIP(), state, ctx.Request().UserAgent())
	sessionStore.Save(googleProvider, sessionID, &session)

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

// TODO A ton a lot of code duplication (~200 lines) from other controllers, get rid of that
// TODO Send user password on creating account using this endpoint (cuz they won't be able to get it somehow else, and won't be able to change password, cuz on changing password of their own account current password also must be included in the request)

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
		return echo.NewHTTPError(
			Error.StatusInternalError.Status(),
			Error.StatusInternalError.Error(),
		)
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
		return echo.NewHTTPError(
			Error.StatusInternalError.Status(),
			Error.StatusInternalError.Error(),
		)
	}

	client := oauth2.NewClient(c, googleOAuthConfig.TokenSource(c, oauthToken))

	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		controller.Log.Error("Failed to get user info from google API", err.Error(), reqMeta)
		return echo.NewHTTPError(
			Error.StatusInternalError.Status(),
			Error.StatusInternalError.Error(),
		)
	}
	defer userInfo.Body.Close()

	body, err := json.Decode[googleUserInfoPartialResponse](userInfo.Body)
	if err != nil {
		return echo.NewHTTPError(
			Error.StatusInternalError.Status(),
			Error.StatusInternalError.Error(),
		)
	}

	if !body.IsEmailVerified {
		errMsg := "Your email isn't verified by google"
		controller.Log.Error("Login failed", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusForbidden, errMsg)
	}

	user, e := DB.Database.GetUserByLogin(body.Email)
	if e != nil && e != Error.StatusNotFound {
		return controller.ConvertErrorStatusToHTTP(e)
	}

	// User with this email doesn't exists -> create new user
	if e == Error.StatusNotFound {
		controller.Log.Trace("Generating password for new user...", reqMeta)

		password, err := pwgen.Generate(&pwgen.Config{
			Length: 32,
			Digits: true,
			Lower: true,
			Upper: true,
		})
		if err != nil {
			errMsg := "Failed to generate password"
			controller.Log.Error("Failed to generate password for new user", errMsg, reqMeta)
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}

		controller.Log.Trace("Generating password for new user: OK", reqMeta)

		uid, e := DB.Database.Create(body.Email, password)
		if e != nil {
			return controller.ConvertErrorStatusToHTTP(e)
		}

		// There are may occur a problem since DB.Database.Create() creates user in primary DB,
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
				return controller.ConvertErrorStatusToHTTP(e)
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
				"The account was created, but it was not possible to log in immediately.",
			)
		}

		payload := &UserDTO.Payload{
			ID: user.ID,
			Login: user.Login,
			Roles: user.Roles,
			SessionID: uuid.NewString(),
			Version: user.Version,
			Audience: []string{config.Auth.SelfAudience}, // TODO Request scopes somehow?
		}

		accessToken, refreshToken, e := token.NewAuthTokens(payload)
		if e != nil {
			return controller.ConvertErrorStatusToHTTP(e)
		}

		session, e := controller.CreateSession(ctx, payload.SessionID, user.ID, config.Auth.RefreshTokenTTL())
		if e != nil {
			return controller.ConvertErrorStatusToHTTP(e)
		}

		if e := DB.Database.SaveSession(session); e != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
		}

		act := ActionDTO.NewUserTargeted(user.ID, user.ID, user.Roles)

		if err := controller.UpdateOrCreateLocation(act, session.ID, session.IpAddress); err != nil {
			if e = DB.Database.RevokeSession(act, session.ID); e != nil {
				return controller.ConvertErrorStatusToHTTP(e)
			}
			return controller.ConvertErrorStatusToHTTP(err)
		}

		ctx.SetCookie(controller.NewAuthCookie(refreshToken))

		return ctx.JSON(
			http.StatusOK,
			ResponseBody.Token{
				Message: "Пользователь успешно создан и авторизован",
				AccessToken: accessToken.String(),
				ExpiresIn: int(accessToken.TTL()) / 1000,
			},
		)
	}

	// If user was found -> log-in already existed user
	deviceID, browser, e := controller.GetDeviceIDAndBrowser(ctx)
	if e != nil {
		return controller.ConvertErrorStatusToHTTP(e)
	}

	// If session with this device is already exists and user tries to login from the same browser
	// (this is guard against cases when auth cookie was lost for some reasons)
	session, e := DB.Database.GetSessionByDeviceAndUserID(deviceID, user.ID)
	if e == nil && session.Browser == browser {
		// TODO code inside this block is very similar with the one that is several lines above, try to fix that
		controller.Log.Info("Already existing user session was found for the specified device. Proceeding with it", reqMeta)

		accessToken, refreshToken, err := controller.UpdateSession(ctx, session, user, &UserDTO.Payload{
			ID: user.ID,
			Login: user.Login,
			Roles: user.Roles,
			SessionID: session.ID,
			Version: user.Version,
		})
		if err == nil {
			ctx.SetCookie(controller.NewAuthCookie(refreshToken))

			controller.Log.Info("Authenticating user '"+body.Email+"': OK", reqMeta)

			return ctx.JSON(
				http.StatusOK,
				ResponseBody.Token{
					Message: "Пользователь успешно авторизован",
					AccessToken: accessToken.String(),
					ExpiresIn: int(accessToken.TTL()) / 1000,
				},
			)
		}
	}

    payload := &UserDTO.Payload{
        ID: user.ID,
        Login: user.Login,
        Roles: user.Roles,
		SessionID: uuid.NewString(),
		Version: user.Version,
		Audience: []string{config.Auth.SelfAudience}, // TODO Request scopes somehow?
    }

    accessToken, refreshToken, e := token.NewAuthTokens(payload)
    if e != nil {
        return controller.ConvertErrorStatusToHTTP(e)
    }

	session, e = controller.CreateSession(ctx, payload.SessionID, user.ID, config.Auth.RefreshTokenTTL())
	if e != nil {
		return controller.ConvertErrorStatusToHTTP(e)
	}

	if err := DB.Database.SaveSession(session); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	act := ActionDTO.NewUserTargeted(user.ID, user.ID, user.Roles)

	if e := controller.UpdateOrCreateLocation(act, session.ID, session.IpAddress); e != nil {
		if e := DB.Database.RevokeSession(act, session.ID); e != nil {
			return controller.ConvertErrorStatusToHTTP(e)
		}
		return controller.ConvertErrorStatusToHTTP(e)
	}

    ctx.SetCookie(controller.NewAuthCookie(refreshToken))

	controller.Log.Info("Authenticating user '"+body.Email+"': OK", reqMeta)

    return ctx.JSON(
        http.StatusOK,
        ResponseBody.Token{
            Message: "Пользователь успешно авторизован",
            AccessToken: accessToken.String(),
            ExpiresIn: int(accessToken.TTL()) / 1000,
        },
    )
}

