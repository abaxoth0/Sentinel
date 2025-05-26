package rolescontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authz"
	controller "sentinel/packages/presentation/api/http/controllers"
	datamodel "sentinel/packages/presentation/data"
	"strings"

	"github.com/labstack/echo/v4"
)

var serviceIdIsNotSpecified = echo.NewHTTPError(
    http.StatusBadRequest,
    "Service ID is not specified",
)

func GetAll(ctx echo.Context) error {
	reqInfo := controller.RequestInfo(ctx)

    serviceID := ctx.Param("serviceID")

    if strings.ReplaceAll(serviceID, " ", "") == "" {
		controller.Logger.Error("Failed to get service roles" + reqInfo, serviceIdIsNotSpecified.Error())
        return serviceIdIsNotSpecified
    }

	controller.Logger.Info("Getting roles for service '"+serviceID+"'..." + reqInfo)

    schema, e := authz.Host.GetSchema(serviceID)
    if e != nil {
		controller.Logger.Error("Failed to get roles of service '"+serviceID+"'" + reqInfo, e.Error())
        return echo.NewHTTPError(http.StatusBadRequest, e.Message)
    }

    roles := make([]string, len(schema.Roles))

    for i, role := range schema.Roles {
        roles[i] = role.Name
    }

	controller.Logger.Info("Getting roles for service '"+serviceID+"': OK" + reqInfo)

    return ctx.JSON(
        http.StatusOK,
        datamodel.RolesResponseBody{ Roles: roles },
    )
}

