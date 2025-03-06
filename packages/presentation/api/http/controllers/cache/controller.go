package cachecontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"

	"github.com/StepanAnanin/weaver"
	"github.com/golang-jwt/jwt"
)

// TODO test
func Drop(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	accessToken, err := token.GetAccessToken(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	filter, err := UserMapper.FilterDTOFromClaims(UserMapper.NoTarget, accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err := authorization.Authorize(authorization.Action.Drop, authorization.Resource.Cache, filter.RequesterRoles); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	cache.Client.Drop()

	res.OK()
}
