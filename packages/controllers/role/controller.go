package role

import (
	"net/http"
	"sentinel/packages/json"
	"sentinel/packages/models/role"
	"sentinel/packages/models/token"
	"sentinel/packages/models/user"
	"sentinel/packages/net"
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

func (c Controller) GetRoles(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodGet); !ok {
		return
	}

	encdoedRoles, ok := json.Encode(role.ListJSON{Roles: role.List})

	if !ok {
		net.Response.InternalServerError(w)

		return
	}

	if err := net.Response.Send(encdoedRoles, w); err != nil {
		net.Request.PrintError("Failed to send OK response", req)
	}

	net.Request.Print("OK", req)
}
