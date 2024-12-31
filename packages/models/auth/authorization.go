package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sentinel/packages/cache"
	ExternalError "sentinel/packages/errs"
	"strings"

	rbac "github.com/StepanAnanin/SentinelRBAC"
)

var Host = (func() *rbac.Host {
	h, e := rbac.LoadHost("RBAC.json")

	if e != nil {
		log.Println(e)
		os.Exit(1)
	}

	return h
})()

var schema = (func() *rbac.Schema {
	s, e := Host.GetSchema("cb663674-803e-4b06-bfeb-87c5cc86383e")

	if e != nil {
		panic(e)
	}

	return s
})()

var user = rbac.NewEntity("user")

type resource struct {
	User  *rbac.Resource
	Cache *rbac.Resource
}

var Resource = &resource{
	User: rbac.NewResource("user", schema.Roles),

	Cache: rbac.NewResource("cache", (func() []*rbac.Role {
		r := []*rbac.Role{}

		for _, role := range schema.Roles {
			// Only admins can interact with cache
			if role.Name == "admin" {
				r = append(r, role)
			} else {
				r = append(r, rbac.NewRole(role.Name, &rbac.Permissions{}))
			}
		}

		return r
	})()),
}

// Checks if user with role == userRoleName can perform action on user with role == targetRoleName.
//
// This method authorize operations only in THIS service!
// Operations on other services must be authorized by themselves!
func Authorize(action rbac.Action, resource *rbac.Resource, userRoles []string) *ExternalError.HTTP {
	cacheKey := fmt.Sprintf("%s[%s->%s]", action.String(), strings.Join(userRoles, ","), resource.Name)
	cacheOK := "K"

	if cacheValue, hit := cache.Get(cacheKey); hit {
		if cacheValue == cacheOK {
			return nil
		}

		return ExternalError.NewHTTP(cacheValue, http.StatusForbidden)
	}

	err := rbac.Authorize(action, resource, userRoles)

	cacheValue := cacheOK

	if err != nil {
		cacheValue = err.Error()
	}

	cache.Set(cacheKey, cacheValue)

	return ExternalError.NewHTTP(err.Error(), http.StatusForbidden)
}
