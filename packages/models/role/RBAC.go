package role

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	externalerror "sentinel/packages/error"
	"sentinel/packages/json"
	"slices"
)

type permission string

const CreatePermission permission = "C"
const SelfCreatePermission permission = "SC"

const ReadPermission permission = "R"
const SelfReadPermission permission = "SR"

const UpdatePermission permission = "U"
const SelfUpdatePermission permission = "SU"

const DeletePermission permission = "D"
const SelfDeletePermission permission = "SD"

const ModeratorPermission permission = "M"
const AdminPermission permission = "A"

var permissions []permission = []permission{
	CreatePermission,
	SelfCreatePermission,
	ReadPermission,
	SelfReadPermission,
	UpdatePermission,
	SelfUpdatePermission,
	DeletePermission,
	SelfDeletePermission,
	ModeratorPermission,
	AdminPermission,
}

type role struct {
	Name        string       `json:"name"`
	Permissions []permission `json:"permissions"`
}

type service struct {
	// uuid format
	ID    string `json:"id"`
	Name  string `json:"name"`
	Roles []role `json:"roles,omitempty"`
}

type rbac struct {
	Roles    []role    `json:"roles"`
	Services []service `json:"services"`
}

// Opens and reads "RBAC.json" file which contains role and permission definitions and returns the parsed configuration.
//
// This function will stop app if it can't read RBAC configuration file, or build RBAC schema.
func loadRBAC() *rbac {
	log.Println("[ RBAC ] Loading configuration...")

	file, err := os.Open("RBAC.json")

	if err != nil {
		if !os.IsExist(err) {
			log.Println("[ CRITICAL ERROR ] RBAC configuration file wasn't found")
			os.Exit(1)
		}

		log.Println(err.Error())
		os.Exit(1)
	}

	// If something will panic, this will be called anyway, unlike if i use this at the end.
	defer func() {
		if err = file.Close(); err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
	}()

	buf, err := io.ReadAll(file)

	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	RBAC, ok := json.Decode[rbac](bytes.NewReader(buf))

	if !ok {
		log.Println("[ CRITICAL ERROR ] Failed to parse RBAC configuration file")
		os.Exit(1)
	}

	log.Println("[ RBAC ] Loading configuration: OK")

	log.Println("[ RBAC ] Checking configuration...")

	if err = checkRBAC(&RBAC); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	log.Println("[ RBAC ] Checking configuration: OK")

	return &RBAC
}

func checkRBAC(RBAC *rbac) error {
	for _, globalRole := range RBAC.Roles {
		for _, permission := range globalRole.Permissions {
			if !slices.Contains(permissions, permission) {
				err := fmt.Sprintf("invalid permission \"%s\" in global role: \"%s\"", string(permission), globalRole.Name)
				return errors.New(err)
			}
		}
	}

	for _, service := range RBAC.Services {
		for _, serviceRole := range service.Roles {
			for _, permission := range serviceRole.Permissions {
				if !slices.Contains(permissions, permission) {
					err := fmt.Sprintf("invalid permission \"%s\" in \"%s\" role: \"%s\"", string(permission), service.Name, serviceRole.Name)
					return errors.New(err)
				}
			}
		}
	}

	return nil
}

var RBAC *rbac = loadRBAC()

func GetServiceRoles(serviceID string) ([]role, *externalerror.Error) {
	var service *service = nil

	for _, rbacService := range RBAC.Services {
		if rbacService.ID == serviceID {
			service = &rbacService
			break
		}
	}

	if service == nil {
		return nil, externalerror.New("service with id \""+serviceID+"\" wasn't found", http.StatusBadRequest)
	}

	roles := []role{}

	// TODO now it's works for O(n**2), try to optimize it.
	// Although it's not so important, RBAC schema isn't big enoungh to see a real difference in performance.
	for _, serviceRole := range service.Roles {
		for _, globalRole := range RBAC.Roles {
			if serviceRole.Name == globalRole.Name {
				roles = append(roles, serviceRole)
			} else {
				roles = append(roles, globalRole)
			}
		}
	}

	return roles, nil
}

func GetAuthRole(roleName string) (*role, *externalerror.Error) {
	for _, globalRole := range RBAC.Roles {
		if globalRole.Name == roleName {
			return &globalRole, nil
		}
	}

	return nil, externalerror.New("role with name \""+roleName+"\" wasn't found", http.StatusBadRequest)
}
