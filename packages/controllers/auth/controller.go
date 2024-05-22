package auth

import (
	"errors"
	"net/http"
	"sentinel/packages/json"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"
	"sentinel/packages/net"
	"strings"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
	If access token expired response will have status 401 (Unauthorized).

	If refresh token expired response will have status 409 (Conflict) also in this case authentication cookie will be deleted.
	Refresh token used only in "Refresh" method. (And mustn't be used in any other function, method, etc.)

	TODO Not sure that status 409 is OK for this case, currently this tells user that there are conflict with server and him,
		 and reason of conflict in next: User assumes that he authorized but it's wrong, cuz refresh token expired.
		 More likely will be better to use status 401 (unathorized) in this case, but once againg - i'm not sure.
*/

type Controller struct {
	user  *user.Model
	token *token.Model
}

func New(dbClient *mongo.Client) *Controller {
	return &Controller{
		user:  user.New(dbClient),
		token: token.New(dbClient),
	}
}

func (c Controller) Login(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPost); !ok {
		return
	}

	body, ok := json.Decode[net.AuthRequestBody](req.Body, w)

	if !ok {
		if err := net.Response.InternalServerError(w); err != nil {
			panic(err)
		}

		return
	}

	iuser, loginError := c.user.Login(body.Email, body.Password)

	if loginError != nil {
		net.Response.Message(loginError.Message, loginError.Status, w)

		net.Request.Print("[ AUTH ERROR ] Invalid auth data.", req)

		return
	}

	accessToken, refreshToken := c.token.Generate(user.Payload{
		ID:    iuser.ID,
		Email: iuser.Email,
		Role:  iuser.Role,
	})

	resBody, ok := json.Encode(net.TokenResponseBody{
		Message:     "Пользователь успешно авторизован",
		AccessToken: accessToken.Value,
	}, w)

	if !ok {
		if err := net.Response.InternalServerError(w); err != nil {
			panic(err)
		}

		return
	}

	http.SetCookie(w, net.Cookie.BuildAuth(refreshToken))

	if err := net.Response.Send(resBody, w); err != nil {
		net.Request.PrintError("Failed to send success response", http.StatusInternalServerError, req)
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

		net.Request.PrintError("Missing refresh token", http.StatusBadRequest, req)

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

	authCookie, err := req.Cookie(token.RefreshTokenKey)

	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			panic(err)
		}

		net.Response.Message("Вы не авторизованы (authentication cookie wasn't found)", http.StatusBadRequest, w)

		net.Request.PrintError("Auth cookie wasn't found", http.StatusBadRequest, req)

		return
	}

	oldRefreshToken, refreshTokenExpired := c.token.ParseRefreshToken(authCookie.Value)

	if !oldRefreshToken.Valid {
		net.Cookie.Delete(authCookie, w)

		net.Response.SendError("Invalid refresh token", http.StatusBadRequest, req, w)

		return
	}

	if refreshTokenExpired {
		net.Cookie.Delete(authCookie, w)

		net.Response.SendError("Refresh token expired", http.StatusConflict, req, w)

		return
	}

	claims := oldRefreshToken.Claims.(jwt.MapClaims)
	email := claims[token.IssuerKey]
	uid := claims[token.SubjectKey]

	if uid == nil || email == nil {
		net.Cookie.Delete(authCookie, w)

		net.Response.Message("Malfunction refresh token", http.StatusBadRequest, w)

		net.Request.PrintError("Malfunction refresh token:\n"+authCookie.Value, http.StatusBadRequest, req)

		return
	}

	accessToken, refreshToken := c.token.Generate(c.token.PayloadFromClaims(claims))

	resBody, ok := json.Encode(net.TokenResponseBody{
		Message:     "Токены успешно обновлены",
		AccessToken: accessToken.Value,
	}, w)

	if !ok {
		if err := net.Response.InternalServerError(w); err != nil {
			panic(err)
		}

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

	authHeaderValue := req.Header.Get("Authorization")

	if authHeaderValue == "" {
		// Now possible error doesn't matter here, but consider it if want to add smth here
		net.Response.Message("Вы не авторизованы", http.StatusUnauthorized, w)

		net.Request.PrintError("Missing authorization header", http.StatusUnauthorized, req)

		return
	}

	accessTokenStr := strings.Split(authHeaderValue, "Bearer ")[1]

	accessToken, accessTokenExpired := c.token.ParseAccessToken(accessTokenStr)

	if !accessToken.Valid {
		net.Response.SendError("Invalid access token", http.StatusBadRequest, req, w)

		return
	}

	if accessTokenExpired {
		net.Response.SendError("Access token expired", http.StatusUnauthorized, req, w)

		return
	}

	claims := accessToken.Claims.(jwt.MapClaims)

	payload := c.token.PayloadFromClaims(claims)

	body, ok := json.Encode(payload, w)

	if !ok {
		if err := net.Response.InternalServerError(w); err != nil {
			panic(err)
		}

		return
	}

	if err := net.Response.Send(body, w); err != nil {
		net.Response.SendError("Failed to send success response", http.StatusInternalServerError, req, w)

		return
	}

	net.Request.Print("Authentication verified for \""+payload.ID+"\".", req)
}
