package role

import (
	"net/http"
	"sentinel/packages/json"
	"sentinel/packages/models/role"
	"sentinel/packages/models/token"
	"sentinel/packages/models/user"

	"github.com/StepanAnanin/weaver/http/response"
	"github.com/StepanAnanin/weaver/logger"
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

func (c *Controller) GetRoles(w http.ResponseWriter, req *http.Request) {
	res := response.New(w)

	encdoedRoles, ok := json.Encode(role.ListJSON{Roles: role.List})

	if !ok {
		res.InternalServerError()

		return
	}

	if err := res.SendBody(encdoedRoles); err != nil {
		logger.PrintError("Failed to send OK response", req)
	}

	logger.Print("OK", req)
}
