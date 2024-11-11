package rolecontroller

import (
	"errors"
	"net/http"
	"sentinel/packages/json"
	"sentinel/packages/models/role"

	"github.com/StepanAnanin/weaver"
)

func GetRoles(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w)

	cookieServiceID, err := req.Cookie("serviceID")

	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			res.BadRequest("Cookie serviceID wasn't found")
			return
		}

		res.InternalServerError()
		return
	}

	serviceID := cookieServiceID.Value

	roles, e := role.GetServiceRoles(serviceID)

	if e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	encdoedRoles, ok := json.Encode(roles)

	if !ok {
		res.InternalServerError()
		return
	}

	if err := res.SendBody(encdoedRoles); err != nil {
		weaver.LogRequestError("Failed to send OK response", req)
	}

	weaver.LogRequest("OK", req)
}
