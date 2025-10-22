package authz

import (
	"fmt"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"strings"

	rbac "github.com/abaxoth0/SentinelRBAC"
)

var log = logger.NewSource("AUTHZ", logger.Default)

var (
	userResource       *rbac.Resource
	cacheResource      *rbac.Resource
	docsResource       *rbac.Resource
	sessionResource    *rbac.Resource
	locationResource   *rbac.Resource
	oauthTokenResource *rbac.Resource
)

var userEntity = rbac.NewEntity("user")

var Schema *rbac.Schema
var Host *rbac.Host

func Init(configPath string) {
	log.Info("Loading configuration file...", nil)

	host, e := rbac.LoadHost(configPath)
	if e != nil {
		log.Fatal("Failed to load configuration file", e.Error(), nil)
	}
	Host = &host

	log.Info("Loading configuration file: OK", nil)
	log.Info("Getting schema for this service...", nil)

	schema, err := Host.GetSchema(config.App.ServiceID)
	if err != nil {
		log.Fatal("Failed to get schema for this service", err.Error(), nil)
	}

	Schema = schema

	log.Info("Getting schema for this service: OK", nil)
	log.Info("Initializing resources...", nil)

	userResource = rbac.NewResource("user")
	cacheResource = rbac.NewResource("cache")
	sessionResource = rbac.NewResource("session")
	locationResource = rbac.NewResource("location")
	oauthTokenResource = rbac.NewResource("oauth_token")
	docsResource = rbac.NewResource("docs")

	log.Info("Initializing resources: OK", nil)

	initContexts()
	initAGP()
}

func stringFromContext(ctx *rbac.AuthorizationContext) string {
	return ctx.Entity.Name() + ":" + ctx.Action.String() + ":" + ctx.Resource.Name()
}

var InsufficientPermissions = Error.NewStatusError(
	"Недостаточно прав для выполнения данной операции",
	http.StatusForbidden,
)

var DeniedByActionGatePolicy = Error.NewStatusError(
	"Authorization has been denied by Action Gate Policy",
	http.StatusForbidden,
)

// Can authorize operations only for the schema of this service.
// Operations on other services must be authorized by themselves!
func authorize(ctx *rbac.AuthorizationContext, rolesNames []string) *Error.Status {
	ctxString := stringFromContext(ctx)

	log.Trace("Authorizing "+ctxString+"...", nil)

	roles := make([]rbac.Role, 0, len(rolesNames))
main_loop:
	for _, roleName := range rolesNames {
		for _, role := range Schema.Roles {
			if roleName == role.Name {
				roles = append(roles, role)
				continue main_loop
			}
		}
		errMsg := "Role " + roleName + " doesn't exist"
		log.Error("Authorization failed", errMsg, nil)
		return Error.NewStatusError(errMsg, http.StatusBadRequest)
	}

	err := rbac.Authorize(ctx, roles, &Schema.ActionGatePolicy)

	if err != nil {
		log.Error("Failed to authorize "+ctxString, err.Error(), nil)

		if err == rbac.InsufficientPermissions {
			return InsufficientPermissions
		}
		switch err {
		case rbac.InsufficientPermissions:
			return InsufficientPermissions
		case rbac.ActionDeniedByAGP:
			return DeniedByActionGatePolicy
		default:
			msg := fmt.Sprintf(
				"Unexpected error occurred during authorization of %s with %s",
				ctxString, strings.Join(rolesNames, ","),
			)
			log.Panic(msg, err.Error(), nil)

			return Error.StatusInternalError
		}
	}

	log.Trace("Authorizing "+ctxString+": OK", nil)

	return nil
}
