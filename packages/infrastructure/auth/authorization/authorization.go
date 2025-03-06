package authorization

import (
	"fmt"
	"log"
	"net/http"
	"os"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/infrastructure/config"
	"strings"

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

// Checks if user with role == userRoleName can perform action on user with role == targetRoleName.
//
// This method authorize operations only in THIS service!
// Operations on other services must be authorized by themselves!
func Authorize(action rbac.Action, resource *rbac.Resource, userRoles []string) *Error.Status {
	cacheKey := fmt.Sprintf("%s[%s->%s]", action.String(), strings.Join(userRoles, ","), resource.Name)
	cacheOK := "K"

	if cacheValue, hit := cache.Client.Get(cacheKey); hit {
		if cacheValue == cacheOK {
			return nil
		}

		return Error.NewStatusError(cacheValue, http.StatusForbidden)
	}

	err := rbac.Authorize(action, resource, userRoles)

	cacheValue := cacheOK

	if err != nil {
		cacheValue = err.Error()
	}

	cache.Client.Set(cacheKey, cacheValue)

	return Error.NewStatusError(err.Error(), http.StatusForbidden)
}

