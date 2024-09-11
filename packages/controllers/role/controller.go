package rolecontroller

import (
	"net/http"
	"sentinel/packages/json"
	"sentinel/packages/models/role"

	"github.com/StepanAnanin/weaver"
)

func GetRoles(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w)

	encdoedRoles, ok := json.Encode(role.ListJSON{Roles: role.List})

	if !ok {
		res.InternalServerError()
		return
	}

	if err := res.SendBody(encdoedRoles); err != nil {
		weaver.LogRequestError("Failed to send OK response", req)
	}

	weaver.LogRequest("OK", req)
}
