package usercontroller

import (
	"log"
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"

	"github.com/StepanAnanin/weaver/http/response"
	"github.com/golang-jwt/jwt"
)

// TODO get rid of code duplication

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

func (c *Controller) buildUserFilter(tartetUID string, req *http.Request) (*user.Filter, *ExternalError.Error) {
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

func (c *Controller) Create(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.AuthRequestBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	_, err := c.user.Create(body.Login, body.Password)

	if err != nil {
		ok, e := ExternalError.Is(err)

		if !ok {
			res.Message("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)

			log.Fatalln(err)
		}

		res.Message(e.Message, e.Status)

		return
	}

	res.OK()
}

func (c *Controller) ChangeLogin(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.UidAndLoginBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	filter, err := c.buildUserFilter(body.UID, req)

	if err != nil {
		res.Message(err.Message, err.Status)
	}

	if e := c.user.ChangeLogin(filter, body.Login); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func (c *Controller) ChangePassword(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.UidAndPasswordBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	filter, err := c.buildUserFilter(body.UID, req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := c.user.ChangePassword(filter, body.Password); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func (c *Controller) ChangeRole(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.UidAndRoleBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	filter, err := c.buildUserFilter(body.UID, req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := c.user.ChangeRole(filter, body.Role); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func (c *Controller) SoftDelete(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.UidBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	filter, err := c.buildUserFilter(body.UID, req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := c.user.SoftDelete(filter); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func (c *Controller) Restore(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.UidBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	filter, err := c.buildUserFilter(body.UID, req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := c.user.Restore(filter); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

// Hard delete
func (c *Controller) Drop(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.UidBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	filter, err := c.buildUserFilter(body.UID, req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err := c.user.Drop(filter); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	res.OK()
}

func (c *Controller) CheckIsLoginExists(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.LoginBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	isExists, err := c.user.CheckIsLoginExists(body.Login)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	resBody, ok := json.Encode(json.LoginExistanceResponseBody{Exists: isExists})

	if !ok {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}

func (c *Controller) GetRole(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	body, ok := json.Decode[json.UidBody](req.Body)

	if !ok {
		res.InternalServerError()
		return
	}

	filter, err := c.buildUserFilter(body.UID, req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	role, err := c.user.GetRole(filter)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	resBody, ok := json.Encode(json.UserRoleResponseBody{Role: role})

	if !ok {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}
