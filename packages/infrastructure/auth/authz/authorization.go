package authz

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"

	rbac "github.com/StepanAnanin/SentinelRBAC"
)

var authzLogger = logger.NewSource("AUTHZ", logger.Default)

type resource struct {
	User  *rbac.Resource
	Cache *rbac.Resource
}

var Resource resource
var Host *rbac.Host
var schema *rbac.Schema

func Init() {
	h, e := rbac.LoadHost("RBAC.json")

	if e != nil {
        authzLogger.Fatal("Failed to load RBAC schema", e.Error())
	}

    Host = &h

	s, err := Host.GetSchema(config.App.ServiceID)

	if err != nil {
		authzLogger.Fatal("Failed to get RBAC schema", err.Error())
	}

    schema = s

    Resource = resource{
        User: rbac.NewResource("user", schema.Roles),

        Cache: rbac.NewResource("cache", (func() []rbac.Role {
            roles := make([]rbac.Role, len(schema.Roles))

            for i, role := range schema.Roles {
                // Only admins can interact with cache
                if role.Name == "admin" {
                    roles[i] = role
                } else {
                    roles[i] = rbac.NewRole(role.Name, 0)
                }
            }

            return roles
        })()),
    }
}

var user = rbac.NewEntity("user")

var insufficientPermissions = Error.NewStatusError(
    "Недостаточно прав для выполнения данной операции",
    http.StatusForbidden,
)

// Checks if user with specified roles can perform action on given resource.
// Returns *Error.Status if user has insufficient permissions or smth is missconfigured, otherwise returns nil.
//
// This method authorize operations only in THIS service!
// Operations on other services must be authorized by themselves!
func Authorize(action rbac.Action, resource *rbac.Resource, userRoles []string) *Error.Status {
	err := resource.Authorize(action, userRoles)

    if err != nil {
        if err == rbac.InsufficientPermissions {
            return insufficientPermissions
        }

        // if err is not nil and not rbac.InsufficientPermissions that means
        // resource permissions wasn't defined for some one of given roles
        // (see rbac.Authorize source code)
        return Error.NewStatusError(
            err.Error(),
            http.StatusInternalServerError,
        )
    }

    return nil
}

