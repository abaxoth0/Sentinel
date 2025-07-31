package authz

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"strings"

	rbac "github.com/StepanAnanin/SentinelRBAC"
)

var log = logger.NewSource("AUTHZ", logger.Default)

var (
	userResource 	*rbac.Resource
	cacheResource	*rbac.Resource
	docsResource	*rbac.Resource
)

var userEntity = rbac.NewEntity("user")

var Schema *rbac.Schema
var Host *rbac.Host

func Init() {
	log.Info("Loading configuration file...", nil)

	h, e := rbac.LoadHost("RBAC.json")
	if e != nil {
        log.Fatal("Failed to load configuration file", e.Error(), nil)
	}

	log.Info("Loading configuration file: OK", nil)
	log.Info("Getting schema for this service...", nil)

    Host = &h

	s, err := Host.GetSchema(config.App.ServiceID)
	if err != nil {
		log.Fatal("Failed to get schema for this service", err.Error(), nil)
	}

	log.Info("Getting schema for this service: OK", nil)
	log.Info("Initializing resources...", nil)

    Schema = s

	userResource = rbac.NewResource("user", Schema.Roles)

	cacheResource = rbac.NewResource("cache", (func() []rbac.Role {
		roles := make([]rbac.Role, len(Schema.Roles))

		for i, role := range Schema.Roles {
			// Only admins can interact with cache
			if role.Name == "admin" {
				roles[i] = role
			} else {
				roles[i] = rbac.NewRole(role.Name, 0)
			}
		}

		return roles
	})())

	docsResource = rbac.NewResource("docs", Schema.Roles)

	log.Info("Initializing resources: OK", nil)
}

var insufficientPermissions = Error.NewStatusError(
    "Недостаточно прав для выполнения данной операции",
    http.StatusForbidden,
)

// TODO is there any point in this function? why just don't use resource.authorize(...)?

// Checks if user with specified roles can perform action on given resource.
// Returns *Error.Status if user has insufficient permissions or smth is missconfigured, otherwise returns nil.
//
// This method authorize operations only in THIS service!
// Operations on other services must be authorized by themselves!
func authorize(action rbac.Action, resource *rbac.Resource, userRoles []string) *Error.Status {
	log.Trace("Authorizing "+action.String()+":"+resource.Name+"...", nil)

	err := resource.Authorize(action, userRoles)

    if err != nil {
		log.Error("Failed to authorize "+action.String()+":"+resource.Name, err.Error(), nil)

        if err == rbac.InsufficientPermissions {
            return insufficientPermissions
        }

		log.Panic(
			"Unexpected error occured during authorization of "+action.String()+":"+resource.Name+" for " + strings.Join(userRoles, ","),
			err.Error(),
			nil,
		)

        // if err is not nil and not rbac.InsufficientPermissions that means
        // resource permissions wasn't defined for some one of given roles
        return Error.NewStatusError(
            err.Error(),
            http.StatusInternalServerError,
        )
    }

	log.Trace("Authorizing "+action.String()+":"+resource.Name+": OK", nil)

    return nil
}

