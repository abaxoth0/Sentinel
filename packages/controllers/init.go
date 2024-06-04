package controllers

import (
	AdminController "sentinel/packages/controllers/admin"
	AuthController "sentinel/packages/controllers/auth"
	RoleController "sentinel/packages/controllers/role"
	UserController "sentinel/packages/controllers/user"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/search"
	"sentinel/packages/models/token"
	"sentinel/packages/models/user"

	"go.mongodb.org/mongo-driver/mongo"
)

type Controllers struct {
	Auth  *AuthController.Controller
	Admin *AdminController.Controller
	User  *UserController.Controller
	Role  *RoleController.Controller
}

func New(dbClient *mongo.Client) *Controllers {
	searchModel := search.New(dbClient)

	userModel := user.New(dbClient, searchModel)
	authModel := auth.New(dbClient, searchModel)

	tokenModel := token.New(dbClient)

	return &Controllers{
		Auth:  AuthController.New(userModel, tokenModel, authModel),
		Admin: AdminController.New(userModel, tokenModel),
		User:  UserController.New(userModel, tokenModel),
		Role:  RoleController.New(userModel, tokenModel),
	}
}
