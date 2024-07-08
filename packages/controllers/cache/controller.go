package cachecontroller

import (
	"net/http"
	"sentinel/packages/cache"
	"sentinel/packages/json"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/token"
	"sentinel/packages/models/user"

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

func (c *Controller) Drop(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	body, ok := json.Decode[map[string]any](req.Body)

	if !ok {
		res.InternalServerError()

		logger.PrintError("Failed to decode JSON", req)

		return
	}

	res.Logged(req)

	accessToken, err := c.token.GetAccessToken(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	filter, err := c.token.UserFilterFromClaims(body["UID"].(string), accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err = filter.RequesterRole.Verify(); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err = auth.Rulebook.DropCache.Authorize(filter.RequesterRole); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	cache.Drop()

	res.OK()
}
