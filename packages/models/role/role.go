package role

import (
	"net/http"
	ExternalError "sentinel/packages/error"
)

const NoneRole string = "none"

func ParseRole(roleName string) (*Role, *ExternalError.Error) {
	for _, rbacRole := range RBAC.Roles {
		if rbacRole.Name == roleName {
			return &rbacRole, nil
		}
	}

	return nil, ExternalError.New("Роль \""+roleName+"\" не надена", http.StatusBadRequest)
}

func GetServiceRoles(serviceID string) ([]Role, *ExternalError.Error) {
	var service *service = nil

	for _, rbacService := range RBAC.Services {
		if rbacService.ID == serviceID {
			service = &rbacService
			break
		}
	}

	if service == nil {
		return nil, ExternalError.New("service with id \""+serviceID+"\" wasn't found", http.StatusBadRequest)
	}

	if len(service.Roles) == 0 {
		return RBAC.Roles, nil
	}

	roles := []Role{}

	// TODO Try to optimize it.
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

// This works only for this service
func GetAuthRole(roleName string) (*Role, *ExternalError.Error) {
	for _, globalRole := range RBAC.Roles {
		if globalRole.Name == roleName {
			return &globalRole, nil
		}
	}

	return nil, ExternalError.New("role with name \""+roleName+"\" wasn't found", http.StatusBadRequest)
}
