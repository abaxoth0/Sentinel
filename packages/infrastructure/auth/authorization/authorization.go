package authorization

import (
	"log"
	"net/http"
	"os"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/config"

	rbac "github.com/StepanAnanin/SentinelRBAC"
)

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
		log.Println(e)
		os.Exit(1)
	}

    Host = h

	s, err := Host.GetSchema(config.Authorization.ServiceID)

	if err != nil {
		panic(err)
	}

    schema = s

    Resource = resource{
        User: rbac.NewResource("user", schema.Roles),

        Cache: rbac.NewResource("cache", (func() []*rbac.Role {
            r := []*rbac.Role{}

            for _, role := range schema.Roles {
                // Only admins can interact with cache
                if role.Name == "admin" {
                    r = append(r, role)
                } else {
                    r = append(r, rbac.NewRole(role.Name, new(rbac.Permissions)))
                }
            }

            return r
        })()),
    }
}

var user = rbac.NewEntity("user")

var insufficientPermissions = Error.NewStatusError(
    "Недостаточно прав для выполнения данной операции",
    http.StatusForbidden,
)

// Checks if user with specified roles can perform action on given resource.
// If can't - returns *Error.Status, nil otherwise.
//
// This method authorize operations only in THIS service!
// Operations on other services must be authorized by themselves!
func Authorize(action rbac.Action, resource *rbac.Resource, userRoles []string) *Error.Status {
	err := rbac.Authorize(action, resource, userRoles)

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

