package rolecontroller

import (
	"net/http"
	"sentinel/packages/json"
	"sentinel/packages/models/role"

	"github.com/StepanAnanin/weaver/http/response"
	"github.com/StepanAnanin/weaver/logger"
)

func GetRoles(w http.ResponseWriter, req *http.Request) {
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
