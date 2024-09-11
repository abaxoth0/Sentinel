package router

import (
	"net/http"
	Auth "sentinel/packages/controllers/auth"
	Cache "sentinel/packages/controllers/cache"
	Role "sentinel/packages/controllers/role"
	User "sentinel/packages/controllers/user"

	"github.com/StepanAnanin/weaver"
	"github.com/gorilla/mux"
)

func Create() *mux.Router {
	router := mux.NewRouter()

	// auth
	router.HandleFunc("/login", weaver.Preprocessing(Auth.Login, http.MethodPost))

	router.HandleFunc("/logout", weaver.Preprocessing(Auth.Logout, http.MethodDelete))

	router.HandleFunc("/refresh", weaver.Preprocessing(Auth.Refresh, http.MethodPut))

	router.HandleFunc("/verify", weaver.Preprocessing(Auth.Verify, http.MethodGet))

	// user
	router.HandleFunc("/user/create", weaver.Preprocessing(User.Create, http.MethodPost))

	router.HandleFunc("/user/delete", weaver.Preprocessing(User.SoftDelete, http.MethodDelete))

	router.HandleFunc("/user/restore", weaver.Preprocessing(User.Restore, http.MethodPut))

	router.HandleFunc("/user/drop", weaver.Preprocessing(User.Drop, http.MethodDelete))

	// TODO Check who can change properties of admin users (no one must do that)
	router.HandleFunc("/user/drop/all-soft-deleted", weaver.Preprocessing(User.DropAllDeleted, http.MethodDelete))

	router.HandleFunc("/user/change/login", weaver.Preprocessing(User.ChangeLogin, http.MethodPatch))

	router.HandleFunc("/user/change/password", weaver.Preprocessing(User.ChangePassword, http.MethodPatch))

	router.HandleFunc("/user/change/role", weaver.Preprocessing(User.ChangeRole, http.MethodPatch))

	router.HandleFunc("/user/check/login", weaver.Preprocessing(User.CheckIsLoginExists, http.MethodPost))

	router.HandleFunc("/user/check/role", weaver.Preprocessing(User.GetRole, http.MethodPost))

	// roles
	router.HandleFunc("/roles", weaver.Preprocessing(Role.GetRoles, http.MethodGet))

	// cache
	router.HandleFunc("/cache/drop", weaver.Preprocessing(Cache.Drop, http.MethodDelete))

	return router
}
