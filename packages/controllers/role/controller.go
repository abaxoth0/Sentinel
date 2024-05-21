package role

import (
	"log"
	"net/http"
	"sentinel/packages/models/token"
	"sentinel/packages/models/user"

	"go.mongodb.org/mongo-driver/mongo"
)

type Controller struct {
	user  *user.Model
	token *token.Model
}

func New(dbClient *mongo.Client) *Controller {
	return &Controller{
		user:  user.New(dbClient),
		token: token.New(dbClient),
	}
}

func (c Controller) GetRoles(w http.ResponseWriter, req *http.Request) {
	log.Fatalln("[ METHOD NOT IMPLEMENTED ]")
}
