package auth

import (
	"errors"
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"

	"github.com/StepanAnanin/weaver/http/cookie"
	"github.com/StepanAnanin/weaver/http/response"
	"github.com/StepanAnanin/weaver/logger"
	"github.com/golang-jwt/jwt"
)

/*

	If access token expired response will have status 401 (Unauthorized).

	If refresh token expired response will have status 409 (Conflict) also in this case authentication cookie will be deleted.
	Refresh token used only in "Refresh" method. (And mustn't be used in any other function, method, etc.)

*/

type Controller struct {
	user  *user.Model
	token *token.Model
	auth  *auth.Model
}

func New(userModel *user.Model, tokenModel *token.Model, authModel *auth.Model) *Controller {
	return &Controller{
		user:  userModel,
		token: tokenModel,
		auth:  authModel,
	}
}

func (c Controller) Login(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	body, ok := json.Decode[json.AuthRequestBody](req.Body)

	if !ok {
		res.InternalServerError()

		return
	}

	iuser, loginError := c.auth.Login(body.Login, body.Password)

	if loginError != nil {
		res.Message(loginError.Message, loginError.Status)

		logger.PrintError("Invalid auth data.", req)

		return
	}

	accessToken, refreshToken := c.token.Generate(&user.Payload{
		ID:    iuser.ID,
		Login: iuser.Login,
		Role:  iuser.Role,
	})

	resBody, ok := json.Encode(json.TokenResponseBody{
		Message:     "Пользователь успешно авторизован",
		AccessToken: accessToken.Value,
	})

	if !ok {
		res.InternalServerError()

		return
	}

	http.SetCookie(w, buildAuthCookie(refreshToken))

	res.SendBody(resBody)

	logger.Print("Authentication successful, user id: "+iuser.ID, req)
}

// This method terminates authentication.
//
// Tokens not used there, cuz it's not matter are they valid or expired, and there are used no methods, that require them.
// Some redundant functional will not change result, it can only add some new prolems. For example:
// User can just stuck, without possibility to logout, cuz this function won't work or will work incorrect.
func (c Controller) Logout(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	authCookie, err := req.Cookie(token.RefreshTokenKey)

	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			panic(err)
		}

		res.Message("Вы не авторизованы (authentication cookie wasn't found)", http.StatusBadRequest)

		logger.PrintError("Missing refresh token", req)

		return
	}

	cookie.Delete(authCookie, w)

	res.OK()

	logger.Print("User logged out.", req)
}

func (c Controller) Refresh(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	oldRefreshToken, e := c.token.GetRefreshToken(req)

	// "e" is either http.ErrNoCookie, either ExternalError.Error
	if e != nil {
		// If auth cookie wasn't found
		if errors.Is(e, http.ErrNoCookie) {
			res.Message("Вы не авторизованы (authentication cookie wasn't found)", http.StatusUnauthorized)

			logger.PrintError("Auth cookie wasn't found", req)
		}

		if isExternal, e := ExternalError.Is(e); isExternal {
			// This cookie used inside of "GetRefreshToken" method so we know that it's exists.
			authCookie, _ := req.Cookie(token.RefreshTokenKey)

			cookie.Delete(authCookie, w)

			res.Message(e.Message, e.Status)

			logger.PrintError(e.Message, req)
		}

		return
	}

	claims := oldRefreshToken.Claims.(jwt.MapClaims)
	payload, err := c.token.PayloadFromClaims(claims)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.Print(err.Message, req)

		return
	}

	accessToken, refreshToken := c.token.Generate(payload)

	resBody, ok := json.Encode(json.TokenResponseBody{
		Message:     "Токены успешно обновлены",
		AccessToken: accessToken.Value,
	})

	if !ok {
		res.InternalServerError()

		return
	}

	http.SetCookie(w, buildAuthCookie(refreshToken))

	res.SendBody(resBody)

	logger.Print("Tokens successfuly refreshed.", req)
}

func (c Controller) Verify(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	accessToken, err := c.token.GetAccessToken(req)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.Print(err.Message, req)

		return
	}

	claims := accessToken.Claims.(jwt.MapClaims)

	payload, err := c.token.PayloadFromClaims(claims)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.Print(err.Message, req)

		return
	}

	body, ok := json.Encode(payload)

	if !ok {
		res.InternalServerError()

		return
	}

	res.SendBody(body)

	logger.Print("Authentication verified for \""+payload.ID+"\".", req)
}
