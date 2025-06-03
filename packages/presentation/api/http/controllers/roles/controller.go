package rolescontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authz"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	datamodel "sentinel/packages/presentation/data"
	"strings"

	"github.com/labstack/echo/v4"
)

var serviceIdIsNotSpecified = echo.NewHTTPError(
    http.StatusBadRequest,
    "Service ID is not specified",
)

func GetAll(ctx echo.Context) error {
	reqMeta, e := request.GetLogMeta(ctx)
	if e != nil {
		controller.Logger.Panic("Failed to get log meta for the request",e.Error(), nil)
		return e
	}

    serviceID := ctx.Param("serviceID")

    if strings.ReplaceAll(serviceID, " ", "") == "" {
		controller.Logger.Error("Failed to get service roles", serviceIdIsNotSpecified.Error(), reqMeta)
        return serviceIdIsNotSpecified
    }

	controller.Logger.Info("Getting roles for service '"+serviceID+"'...", reqMeta)

    schema, err := authz.Host.GetSchema(serviceID)
    if err != nil {
		controller.Logger.Error("Failed to get roles of service '"+serviceID+"'", err.Error(), reqMeta)
        return echo.NewHTTPError(http.StatusBadRequest, err.Message)
    }

    roles := make([]string, len(schema.Roles))

    for i, role := range schema.Roles {
        roles[i] = role.Name
    }

	controller.Logger.Info("Getting roles for service '"+serviceID+"': OK", reqMeta)

    return ctx.JSON(
        http.StatusOK,
        datamodel.RolesResponseBody{ Roles: roles },
    )
}

