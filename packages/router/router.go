package router

import (
	"sentinel/packages/config"
	"sentinel/packages/controllers"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// Defines endpoints and applying handlers to them. Also creates all controllers.
func Init(dbClient *mongo.Client) *mux.Router {
	router := mux.NewRouter()

	controller := controllers.New(dbClient)

	// auth
	router.HandleFunc("/login", controller.Auth.Login)
	router.HandleFunc("/logout", controller.Auth.Logout)
	router.HandleFunc("/refresh", controller.Auth.Refresh)
	router.HandleFunc("/verify", controller.Auth.Verify)

	// user
	router.HandleFunc("/user/create", controller.User.Create)
	router.HandleFunc("/user/delete", controller.User.SoftDelete)
	router.HandleFunc("/user/restore", controller.User.Restore)
	router.HandleFunc("/user/drop", controller.User.Drop)
	router.HandleFunc("/user/change/email", controller.User.ChangeEmail)
	router.HandleFunc("/user/change/password", controller.User.ChangePassword)
	router.HandleFunc("/user/change/role", controller.User.ChangeRole)

	// roles
	router.HandleFunc("/roles", controller.Role.GetRoles)

	if config.Debug.Enabled {

		router.HandleFunc("/admin/clear-cache", controller.Admin.DropCache)
	}

	return router
}
