package auth

import (
	"errors"
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"
	"sentinel/packages/net"

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
	if ok := net.Request.Preprocessing(w, req, http.MethodPost); !ok {
		return
	}

	body, ok := json.Decode[net.AuthRequestBody](req.Body)

	if !ok {
		net.Response.InternalServerError(w)

		return
	}

	iuser, loginError := c.auth.Login(body.Email, body.Password)

	if loginError != nil {
		net.Response.Message(loginError.Message, loginError.Status, w)

		net.Request.Print("[ AUTH ERROR ] Invalid auth data.", req)

		return
	}

	accessToken, refreshToken := c.token.Generate(&user.Payload{
		ID:    iuser.ID,
		Email: iuser.Email,
		Role:  iuser.Role,
	})

	resBody, ok := json.Encode(net.TokenResponseBody{
		Message:     "Пользователь успешно авторизован",
		AccessToken: accessToken.Value,
	})

	if !ok {
		net.Response.InternalServerError(w)

		return
	}

	http.SetCookie(w, net.Cookie.BuildAuth(refreshToken))

	if err := net.Response.Send(resBody, w); err != nil {
		net.Request.PrintError("Failed to send success response", req)
	}

	net.Request.Print("Authentication successful, user id: "+iuser.ID, req)
}

// This method terminates authentication.
//
// Tokens not used there, cuz it's not matter are they valid or expired, and there are used no methods, that require them.
// Some redundant functional will not change result, it can only add some new prolems. For example:
// User can just stuck, without possibility to logout, cuz this function won't work or will work incorrect.
func (c Controller) Logout(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	authCookie, err := req.Cookie(token.RefreshTokenKey)

	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			panic(err)
		}

		net.Response.Message("Вы не авторизованы (authentication cookie wasn't found)", http.StatusBadRequest, w)

		net.Request.PrintError("Missing refresh token", req)

		return
	}

	net.Cookie.Delete(authCookie, w)

	// In error case log already done in `net.SendOkResponse`
	if err := net.Response.OK(w); err != nil {
		return
	}

	net.Request.Print("User logged out.", req)
}

func (c Controller) Refresh(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPut); !ok {
		return
	}

	oldRefreshToken, e := c.token.GetRefreshToken(req)

	// "e" is either http.ErrNoCookie, either ExternalError.Error
	if e != nil {
		// If auth cookie wasn't found
		if errors.Is(e, http.ErrNoCookie) {
			net.Response.Message("Вы не авторизованы (authentication cookie wasn't found)", http.StatusUnauthorized, w)

			net.Request.PrintError("Auth cookie wasn't found", req)
		}

		if isExternal, e := ExternalError.Is(e); isExternal {
			// This cookie used inside of "GetRefreshToken" method so we know that it's exists.
			authCookie, _ := req.Cookie(token.RefreshTokenKey)

			net.Cookie.Delete(authCookie, w)

			net.Response.SendError(e.Message, e.Status, req, w)
		}

		return
	}

	claims := oldRefreshToken.Claims.(jwt.MapClaims)
	payload, err := c.token.PayloadFromClaims(claims)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	accessToken, refreshToken := c.token.Generate(payload)

	resBody, ok := json.Encode(net.TokenResponseBody{
		Message:     "Токены успешно обновлены",
		AccessToken: accessToken.Value,
	})

	if !ok {
		net.Response.InternalServerError(w)

		return
	}

	http.SetCookie(w, net.Cookie.BuildAuth(refreshToken))

	if err := net.Response.Send(resBody, w); err != nil {
		net.Response.SendError("Failed to send success response", http.StatusInternalServerError, req, w)
	}

	net.Request.Print("Tokens successfuly refreshed.", req)
}

func (c Controller) Verify(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodGet); !ok {
		return
	}

	accessToken, err := c.token.GetAccessToken(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	claims := accessToken.Claims.(jwt.MapClaims)

	payload, err := c.token.PayloadFromClaims(claims)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	body, ok := json.Encode(payload)

	if !ok {
		net.Response.InternalServerError(w)

		return
	}

	if err := net.Response.Send(body, w); err != nil {
		net.Response.SendError("Failed to send success response", http.StatusInternalServerError, req, w)

		return
	}

	net.Request.Print("Authentication verified for \""+payload.ID+"\".", req)
}
