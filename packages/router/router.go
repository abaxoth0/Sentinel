package router

import (
	"sentinel/packages/config"
	admincontroller "sentinel/packages/controllers/admin"
	authcontroller "sentinel/packages/controllers/auth"
	rolecontroller "sentinel/packages/controllers/role"
	usercontroller "sentinel/packages/controllers/user"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// Defines endpoints and applying handlers to them. Also creates all controllers.
func Init(dbClient *mongo.Client) *mux.Router {
	router := mux.NewRouter()

	authController := authcontroller.New(dbClient)
	userController := usercontroller.New(dbClient)
	adminController := admincontroller.New(dbClient)
	roleController := rolecontroller.New(dbClient)

	router.HandleFunc("/login", authController.Login)
	router.HandleFunc("/logout", authController.Logout)
	router.HandleFunc("/refresh", authController.Refresh)
	router.HandleFunc("/verification", authController.Verify)

	router.HandleFunc("/user/create", userController.Create)

	router.HandleFunc("/roles", roleController.GetRoles)

	if config.Debug.Enabled {
		// TODO implement all
		router.HandleFunc("/user/delete", userController.UNSAFE_SoftDelete)
		router.HandleFunc("/user/drop", userController.UNSAFE_HardDelete)
		router.HandleFunc("/user/change/email", userController.UNSAFE_ChangeEmail)
		router.HandleFunc("/user/change/password", userController.UNSAFE_ChangePassword)
		router.HandleFunc("/user/change/role", userController.UNSAFE_ChangeRole)

		router.HandleFunc("/admin/clear-cache", adminController.DropCache)
	}

	return router
}
