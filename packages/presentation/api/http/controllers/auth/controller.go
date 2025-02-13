package authcontroller

import (
	"errors"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errs"
	"sentinel/packages/infrastructure/auth/authentication"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	datamodel "sentinel/packages/presentation/data"

	"github.com/StepanAnanin/weaver"
	"github.com/golang-jwt/jwt"
)

/*

	If access token expired response will have status 401 (Unauthorized).

	If refresh token expired response will have status 409 (Conflict) also in this case authentication cookie will be deleted.
	Refresh token used only in "Refresh" method. (And mustn't be used in any other function, method, et)

*/

func Login(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	body, err := datamodel.Decode[datamodel.AuthRequestBody](req.Body)

	if err != nil {
		res.InternalServerError()
		return
	}

	user, loginError := authentication.Login(body.Login, body.Password)

	if loginError != nil {
		res.Message(loginError.Message, loginError.Status)
		return
	}

	accessToken, refreshToken := token.Generate(&UserDTO.Payload{
		ID:    user.ID,
		Login: user.Login,
		Roles: user.Roles,
	})

	resBody, err := datamodel.Encode(datamodel.TokenResponseBody{
		Message:     "Пользователь успешно авторизован",
		AccessToken: accessToken.Value,
	})

	if err != nil {
		res.InternalServerError()
		return
	}

	http.SetCookie(w, buildAuthCookie(refreshToken))

	res.SendBody(resBody)

	weaver.LogRequest("Authentication successful, user id: "+user.ID, req)
}

func Logout(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w)

	authCookie, err := req.Cookie(token.RefreshTokenKey)

	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			panic(err)
		}

		res.Message("Вы не авторизованы (authentication cookie wasn't found)", http.StatusBadRequest)

		weaver.LogRequest("Missing refresh token", req)

		return
	}

	weaver.DeleteCookie(authCookie, w)

	res.OK()

	weaver.LogRequest("User logged out.", req)
}

func Refresh(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w)

	oldRefreshToken, e := token.GetRefreshToken(req)

	// "e" is either http.ErrNoCookie, either Error.Status
	if e != nil {
		// If auth cookie wasn't found
		if errors.Is(e, http.ErrNoCookie) {
			res.Message("Вы не авторизованы (authentication cookie wasn't found)", http.StatusUnauthorized)

			weaver.LogRequest("Auth cookie wasn't found", req)
		}

		if is, e := Error.IsStatusError(e); is {
			// This cookie used inside of "GetRefreshToken" method so we know that it's exists.
			authCookie, _ := req.Cookie(token.RefreshTokenKey)

			weaver.DeleteCookie(authCookie, w)

			res.Message(e.Message, e.Status)

			weaver.LogRequest(e.Message, req)
		}

		return
	}

	claims := oldRefreshToken.Claims.(jwt.MapClaims)
	payload, err := UserMapper.PayloadFromClaims(claims)

	if err != nil {
		res.Message(err.Message, err.Status)

		weaver.LogRequest(err.Message, req)

		return
	}

	accessToken, refreshToken := token.Generate(payload)

	resBody, er := datamodel.Encode(datamodel.TokenResponseBody{
		Message:     "Токены успешно обновлены",
		AccessToken: accessToken.Value,
	})

	if er != nil {
		res.InternalServerError()
		return
	}

	http.SetCookie(w, buildAuthCookie(refreshToken))

	res.SendBody(resBody)

	weaver.LogRequest("Tokens successfuly refreshed.", req)
}

func Verify(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	accessToken, err := token.GetAccessToken(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	claims := accessToken.Claims.(jwt.MapClaims)

	payload, err := UserMapper.PayloadFromClaims(claims)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	body, e := datamodel.Encode(payload)

	if e != nil {
		res.InternalServerError()
		return
	}

	res.SendBody(body)

	weaver.LogRequest("Authentication verified for \""+payload.ID+"\".", req)
}
