package router

import (
	"net/http"
	"sentinel/packages/config"
	"sentinel/packages/controllers"

	"github.com/StepanAnanin/weaver/http/request"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// Defines endpoints and applying handlers to them. Also creates all controllers.
func Init(dbClient *mongo.Client) *mux.Router {
	router := mux.NewRouter()

	controller := controllers.New(dbClient)

	// auth
	router.HandleFunc("/login", request.Preprocessing(controller.Auth.Login, []string{http.MethodPost}))

	router.HandleFunc("/logout", request.Preprocessing(controller.Auth.Logout, []string{http.MethodDelete}))

	router.HandleFunc("/refresh", request.Preprocessing(controller.Auth.Refresh, []string{http.MethodPut}))

	router.HandleFunc("/verify", request.Preprocessing(controller.Auth.Verify, []string{http.MethodGet}))

	// user
	router.HandleFunc("/user/create", request.Preprocessing(controller.User.Create, []string{http.MethodPost}))

	router.HandleFunc("/user/delete", request.Preprocessing(controller.User.SoftDelete, []string{http.MethodDelete}))

	router.HandleFunc("/user/restore", request.Preprocessing(controller.User.Restore, []string{http.MethodPut}))

	router.HandleFunc("/user/drop", request.Preprocessing(controller.User.Drop, []string{http.MethodPost}))

	router.HandleFunc("/user/change/login", request.Preprocessing(controller.User.ChangeLogin, []string{http.MethodPatch}))

	router.HandleFunc("/user/change/password", request.Preprocessing(controller.User.ChangePassword, []string{http.MethodPatch}))

	router.HandleFunc("/user/change/role", request.Preprocessing(controller.User.ChangeRole, []string{http.MethodPatch}))

	// TODO Add slug and get login from here instead of a request body
	router.HandleFunc("/user/check/login", request.Preprocessing(controller.User.CheckIsLoginExists, []string{http.MethodPost}))

	// roles
	router.HandleFunc("/roles", request.Preprocessing(controller.Role.GetRoles, []string{http.MethodPatch}))

	if config.Debug.Enabled {
		router.HandleFunc("/admin/clear-cache", request.Preprocessing(controller.Admin.DropCache, []string{http.MethodDelete}))
	}

	return router
}
