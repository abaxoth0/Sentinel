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
func (c Controller) buildReqBodyAndUserFilter(w http.ResponseWriter, req *http.Request) (map[string]any, *user.Filter, *ExternalError.Error) {
	// empty body
	var emptyBody map[string]any
	// empty token
	var emptyFilter *user.Filter

	accessToken, err := c.token.GetAccessToken(req)

	if err != nil {
		return emptyBody, emptyFilter, err
	}

	body, ok := json.Decode[map[string]any](req.Body, w)

	if !ok {
		return emptyBody, emptyFilter, ExternalError.New("Internal Server Error", http.StatusInternalServerError)
	}

	// If token is valid, then we can trust claims
	filter, err := c.token.UserFilterFromClaims(body["UID"].(string), accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		return emptyBody, emptyFilter, err
	}

	if err := filter.RequesterRole.Verify(); err != nil {
		return emptyBody, emptyFilter, err
	}

	return body, filter, nil
}

func (c Controller) changeUserProperty(targetProperty property, allowedMethod string, w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, allowedMethod); !ok {
		return
	}

	if ok := targetProperty.Verify(); !ok {
		net.Response.SendError("Invalid user property", http.StatusInternalServerError, req, w)

		return
	}

	body, filter, err := c.buildReqBodyAndUserFilter(w, req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	var e *ExternalError.Error

	switch targetProperty {
	case emailProperty:
		e = c.user.ChangeEmail(filter, body["email"].(string))
	case passwordProperty:
		e = c.user.ChangePassword(filter, body["password"].(string))
	case roleProperty:
		e = c.user.ChangeRole(filter, body["role"].(string))
	case deletedAtProperty:
		if allowedMethod == http.MethodDelete {
			e = c.user.SoftDelete(filter)

			break
		}

		if allowedMethod == http.MethodPut {
			e = c.user.Restore(filter)

			break
		}

		e = ExternalError.New("Invalid request method", http.StatusBadRequest)
	default:
		net.Request.PrintError("Invalid user property", 500, req)
	}

	if e != nil {
		net.Response.SendError(e.Message, e.Status, req, w)

		return
	}

	net.Response.OK(w)
}

func (c Controller) UNSAFE_ChangeEmail(w http.ResponseWriter, req *http.Request) {
	c.changeUserProperty(emailProperty, http.MethodPatch, w, req)
}

func (c Controller) UNSAFE_ChangePassword(w http.ResponseWriter, req *http.Request) {
	c.changeUserProperty(passwordProperty, http.MethodPatch, w, req)
}

func (c Controller) UNSAFE_ChangeRole(w http.ResponseWriter, req *http.Request) {
	c.changeUserProperty(roleProperty, http.MethodPatch, w, req)
}

func (c Controller) SoftDelete(w http.ResponseWriter, req *http.Request) {
	c.changeUserProperty(deletedAtProperty, http.MethodDelete, w, req)
}

func (c Controller) Restore(w http.ResponseWriter, req *http.Request) {
	c.changeUserProperty(deletedAtProperty, http.MethodPut, w, req)
}

func (c Controller) UNSAFE_HardDelete(w http.ResponseWriter, req *http.Request) {
	net.Response.InternalServerError(w)

	log.Fatalln("[ CRITICAL ERROR ] Method not implemented")
}
