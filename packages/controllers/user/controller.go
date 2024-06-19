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

// Retrieves untyped request body from given request
func (c Controller) buildReqBody(req *http.Request) (map[string]any, *ExternalError.Error) {
	body, ok := json.Decode[map[string]any](req.Body)

	if !ok {
		return map[string]any{}, ExternalError.New("Internal Server Error (failed to decode JSON)", http.StatusInternalServerError)
	}

	return body, nil
}

func (c Controller) buildUserFilter(tartetUID string, req *http.Request) (*user.Filter, *ExternalError.Error) {
	var emptyFilter *user.Filter

	accessToken, err := c.token.GetAccessToken(req)

	if err != nil {
		return emptyFilter, err
	}

	// If token is valid, then we can trust claims
	filter, err := c.token.UserFilterFromClaims(tartetUID, accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		return emptyFilter, err
	}

	if err := filter.RequesterRole.Verify(); err != nil {
		return emptyFilter, err
	}

	return filter, nil
}

func (c Controller) getRequestBodyAndUserFilter(req *http.Request) (map[string]any, *user.Filter, *ExternalError.Error) {
	body, bodyErr := c.buildReqBody(req)
	filter, filterErr := c.buildUserFilter(body["UID"].(string), req)

	if bodyErr != nil {
		return map[string]any{}, &user.Filter{}, bodyErr
	}

	if filterErr != nil {
		return map[string]any{}, &user.Filter{}, filterErr
	}

	return body, filter, nil
}

func (c Controller) Create(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPost); !ok {
		return
	}

	body, ok := json.Decode[net.AuthRequestBody](req.Body)

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

		net.Request.PrintError("Failed to create new user: "+e.Message, req)

		return
	}

	if err := net.Response.OK(w); err != nil {
		net.Response.SendError("Failed to send success response", http.StatusInternalServerError, req, w)

		return
	}

	net.Request.Print("New user created, email: "+body.Email, req)
}

func (c Controller) ChangeEmail(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPatch); !ok {
		return
	}

	body, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	if e := c.user.ChangeEmail(filter, body["email"].(string)); e != nil {
		net.Response.SendError(e.Message, e.Status, req, w)
	}
}

func (c Controller) ChangePassword(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPatch); !ok {
		return
	}

	body, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	if e := c.user.ChangePassword(filter, body["password"].(string)); e != nil {
		net.Response.SendError(e.Message, e.Status, req, w)
	}
}

func (c Controller) ChangeRole(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPatch); !ok {
		return
	}

	body, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	if e := c.user.ChangeRole(filter, body["role"].(string)); e != nil {
		net.Response.SendError(e.Message, e.Status, req, w)
	}
}

func (c Controller) SoftDelete(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	_, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	if e := c.user.SoftDelete(filter); e != nil {
		net.Response.SendError(e.Message, e.Status, req, w)
	}
}

func (c Controller) Restore(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPut); !ok {
		return
	}

	_, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	if e := c.user.Restore(filter); e != nil {
		net.Response.SendError(e.Message, e.Status, req, w)
	}
}

// Hard delete
func (c Controller) Drop(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	body, bodyErr := c.buildReqBody(req)
	filter, filterErr := c.buildUserFilter(body["UID"].(string), req)

	if bodyErr != nil || filterErr != nil {
		var err *ExternalError.Error

		if bodyErr != nil {
			err = bodyErr
		} else {
			err = filterErr
		}

		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	if err := c.user.Drop(filter); err != nil {
		net.Response.Message(err.Message, err.Status, w)

		net.Request.PrintError(err.Message, req)

		return
	}

	net.Response.OK(w)
}
