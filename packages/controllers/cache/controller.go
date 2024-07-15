package cachecontroller

import (
	"net/http"
	"sentinel/packages/cache"
	"sentinel/packages/models/role"
	"sentinel/packages/models/token"

	"github.com/StepanAnanin/weaver/http/response"
	"github.com/golang-jwt/jwt"
)

func Drop(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	accessToken, err := token.GetAccessToken(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	// There is no targeted user, so just pass empty string
	filter, err := token.UserFilterFromClaims("", accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err = filter.RequesterRole.Verify(); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if filter.RequesterRole != role.Administrator {
		res.Message("Недостаточно прав для выполнения данной операции", http.StatusForbidden)
		return
	}

	cache.Drop()

	res.OK()
}
