package rolescontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authorization"
	datamodel "sentinel/packages/presentation/data"

	"github.com/labstack/echo/v4"
)

var serviceIsNotSpecified = echo.NewHTTPError(
    http.StatusBadRequest,
    "Service isn't specified (missing serviceID cookie)",
)

func GetAll(ctx echo.Context) error {
    serviceID := ctx.Param("serviceID")

    if serviceID == "" {
        return serviceIsNotSpecified
    }

    schema, e := authorization.Host.GetSchema(serviceID)

    if e != nil {
        return echo.NewHTTPError(http.StatusBadRequest, e.Message)
    }

    roles := make([]string, len(schema.Roles))

    for i, role := range schema.Roles {
        roles[i] = role.Name
    }

    return ctx.JSON(
        http.StatusOK,
        datamodel.RolesResponseBody{ Roles: roles },
    )
}

