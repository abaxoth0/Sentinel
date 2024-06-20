package user

import (
	"log"
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"

	"github.com/StepanAnanin/weaver/http/response"
	"github.com/StepanAnanin/weaver/logger"
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
	res := response.New(w)

	body, ok := json.Decode[json.AuthRequestBody](req.Body)

	if !ok {
		res.InternalServerError()

		return
	}

	_, err := c.user.Create(body.Email, body.Password)

	if err != nil {
		ok, e := ExternalError.Is(err)

		if !ok {
			res.Message("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)

			log.Fatalln(err)
		}

		res.Message(e.Message, e.Status)

		logger.PrintError("Failed to create new user: "+e.Message, req)

		return
	}

	res.OK()

	logger.Print("New user created, email: "+body.Email, req)
}

func (c Controller) ChangeEmail(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	body, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.PrintError(err.Message, req)

		return
	}

	if e := c.user.ChangeEmail(filter, body["email"].(string)); e != nil {
		res.Message(e.Message, e.Status)

		logger.PrintError(e.Message, req)
	}
}

func (c Controller) ChangePassword(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	body, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.PrintError(err.Message, req)

		return
	}

	if e := c.user.ChangePassword(filter, body["password"].(string)); e != nil {
		res.Message(e.Message, e.Status)

		logger.PrintError(e.Message, req)
	}
}

func (c Controller) ChangeRole(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	body, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.PrintError(err.Message, req)

		return
	}

	if e := c.user.ChangeRole(filter, body["role"].(string)); e != nil {
		res.Message(e.Message, e.Status)

		logger.PrintError(e.Message, req)
	}
}

func (c Controller) SoftDelete(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	_, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.PrintError(err.Message, req)

		return
	}

	if e := c.user.SoftDelete(filter); e != nil {
		res.Message(e.Message, e.Status)

		logger.PrintError(e.Message, req)
	}
}

func (c Controller) Restore(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	_, filter, err := c.getRequestBodyAndUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)

		logger.PrintError(err.Message, req)

		return
	}

	if e := c.user.Restore(filter); e != nil {
		res.Message(e.Message, e.Status)

		logger.PrintError(e.Message, req)
	}
}

// Hard delete
func (c Controller) Drop(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	body, bodyErr := c.buildReqBody(req)
	filter, filterErr := c.buildUserFilter(body["UID"].(string), req)

	if bodyErr != nil || filterErr != nil {
		var err *ExternalError.Error

		if bodyErr != nil {
			err = bodyErr
		} else {
			err = filterErr
		}

		res.Message(err.Message, err.Status)

		logger.PrintError(err.Message, req)

		return
	}

	if err := c.user.Drop(filter); err != nil {
		res.Message(err.Message, err.Status)

		logger.PrintError(err.Message, req)

		return
	}

	res.OK()
}
