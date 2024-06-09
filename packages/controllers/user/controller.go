package user

import (
	"log"
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"
	"sentinel/packages/net"

	"github.com/golang-jwt/jwt"
)

type Controller struct {
	user  *user.Model
	token *token.Model
}

func New(userModel *user.Model, tokenModel *token.Model) *Controller {
	return &Controller{
		user:  userModel,
		token: tokenModel,
	}
}

func (c Controller) Create(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPost); !ok {
		return
	}

	body, ok := json.Decode[net.AuthRequestBody](req.Body, w)

	if !ok {
		net.Response.InternalServerError(w)

		return
	}

	_, err := c.user.Create(body.Email, body.Password)

	if err != nil {
		ok, e := ExternalError.Is(err)

		if !ok {
			net.Response.Message("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError, w)

			log.Fatalln(err)
		}

		net.Response.Message(e.Message, e.Status, w)

		net.Request.PrintError("Failed to create new user: "+e.Message, e.Status, req)

		return
	}

	if err := net.Response.OK(w); err != nil {
		net.Response.SendError("Failed to send success response", http.StatusInternalServerError, req, w)

		return
	}

	net.Request.Print("New user created, email: "+body.Email, req)
}

// Returns untyped request body, access token and true if no errors occurred, false otherwise
func (c Controller) buildReqBodyAndUserFilter(w http.ResponseWriter, req *http.Request) (interface{}, *user.Filter, bool) {
	// empty body
	var emptyBody any
	// empty token
	var emptyFilter *user.Filter

	accessToken, err := c.token.GetAccessToken(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return emptyBody, emptyFilter, false
	}

	body, ok := json.Decode[net.UidBody](req.Body, w)

	if !ok {
		if err := net.Response.InternalServerError(w); err != nil {
			panic(err)
		}
	}

	// If token is valid, then we can trust claims
	filter, err := c.token.UserFilterFromClaims(body.UID, accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return emptyBody, emptyFilter, false
	}

	if err := filter.RequesterRole.Verify(); err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return emptyBody, emptyFilter, false
	}

	return body, filter, true
}

func (c Controller) UNSAFE_ChangeEmail(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	rawBody, filter, ok := c.buildReqBodyAndUserFilter(w, req)

	// Response was already sent
	if !ok {
		return
	}

	body := rawBody.(net.UidAndEmailBody)

	if err := c.user.ChangeEmail(filter, body.Email); err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	net.Response.OK(w)
}

func (c Controller) UNSAFE_ChangePassword(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	rawBody, filter, ok := c.buildReqBodyAndUserFilter(w, req)

	// Response was already sent
	if !ok {
		return
	}

	body := rawBody.(net.UidAndPasswordBody)

	if err := c.user.ChangePassword(filter, body.Password); err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	net.Response.OK(w)
}

func (c Controller) UNSAFE_ChangeRole(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	rawBody, filter, ok := c.buildReqBodyAndUserFilter(w, req)

	// Response was already sent
	if !ok {
		return
	}

	body := rawBody.(net.UidAndRoleBody)

	if err := c.user.ChangeRole(filter, body.Role); err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	net.Response.OK(w)
}

func (c Controller) SoftDelete(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	_, filter, ok := c.buildReqBodyAndUserFilter(w, req)

	// Response was already sent
	if !ok {
		return
	}

	if err := c.user.SoftDelete(filter); err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	net.Response.OK(w)
}

func (c Controller) UNSAFE_HardDelete(w http.ResponseWriter, req *http.Request) {
	net.Response.InternalServerError(w)

	log.Fatalln("[ CRITICAL ERROR ] Method not implemented")
}

func (c Controller) Restore(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	_, filter, ok := c.buildReqBodyAndUserFilter(w, req)

	// Response was already sent
	if !ok {
		return
	}

	if err := c.user.Restore(filter); err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	net.Response.OK(w)
}
