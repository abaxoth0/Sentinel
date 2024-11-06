package cachecontroller

import (
	"net/http"
	"sentinel/packages/cache"
	"sentinel/packages/models/token"

	"github.com/StepanAnanin/weaver"
	"github.com/golang-jwt/jwt"
)

func Drop(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

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

	// TODO temporary solution
	if filter.RequesterRole != "admin" {
		res.Forbidden("Only admin can drop cache.")
		return
	}

	cache.Drop()

	res.OK()
}
