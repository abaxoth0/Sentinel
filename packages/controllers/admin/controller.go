package admin

import (
	"log"
	"net/http"
	"sentinel/packages/models/token"
	"sentinel/packages/models/user"
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

func (c Controller) DropCache(w http.ResponseWriter, req *http.Request) {
	log.Fatalln("[ METHOD NOT IMPLEMENTED ]")
}
