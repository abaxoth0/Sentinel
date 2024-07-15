package router

import (
	"net/http"
	Auth "sentinel/packages/controllers/auth"
	Cache "sentinel/packages/controllers/cache"
	Role "sentinel/packages/controllers/role"
	User "sentinel/packages/controllers/user"

	"github.com/StepanAnanin/weaver/http/request"
	"github.com/gorilla/mux"
)

// Defines endpoints and applying handlers to them. Also creates all controllers.
func Init() *mux.Router {
	router := mux.NewRouter()

	// auth
	router.HandleFunc("/login", request.Preprocessing(Auth.Login, []string{http.MethodPost}))

	router.HandleFunc("/logout", request.Preprocessing(Auth.Logout, []string{http.MethodDelete}))

	router.HandleFunc("/refresh", request.Preprocessing(Auth.Refresh, []string{http.MethodPut}))

	router.HandleFunc("/verify", request.Preprocessing(Auth.Verify, []string{http.MethodGet}))

	// user
	router.HandleFunc("/user/create", request.Preprocessing(User.Create, []string{http.MethodPost}))

	router.HandleFunc("/user/delete", request.Preprocessing(User.SoftDelete, []string{http.MethodDelete}))

	router.HandleFunc("/user/restore", request.Preprocessing(User.Restore, []string{http.MethodPut}))

	router.HandleFunc("/user/drop", request.Preprocessing(User.Drop, []string{http.MethodPost}))

	router.HandleFunc("/user/change/login", request.Preprocessing(User.ChangeLogin, []string{http.MethodPatch}))

	router.HandleFunc("/user/change/password", request.Preprocessing(User.ChangePassword, []string{http.MethodPatch}))

	router.HandleFunc("/user/change/role", request.Preprocessing(User.ChangeRole, []string{http.MethodPatch}))

	router.HandleFunc("/user/check/login", request.Preprocessing(User.CheckIsLoginExists, []string{http.MethodPost}))

	router.HandleFunc("/user/check/role", request.Preprocessing(User.GetRole, []string{http.MethodPost}))

	// roles
	router.HandleFunc("/roles", request.Preprocessing(Role.GetRoles, []string{http.MethodGet}))

	// cache
	router.HandleFunc("/cache/drop", request.Preprocessing(Cache.Drop, []string{http.MethodDelete}))

	return router
}
