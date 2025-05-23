package rolescontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authz"
	datamodel "sentinel/packages/presentation/data"
	"strings"

	"github.com/labstack/echo/v4"
)

var serviceIsMissing = echo.NewHTTPError(
    http.StatusBadRequest,
    "Service ID is missing",
)

func GetAll(ctx echo.Context) error {
    serviceID := ctx.Param("serviceID")

    if strings.ReplaceAll(serviceID, " ", "") == "" {
        return serviceIsMissing
    }

    schema, e := authz.Host.GetSchema(serviceID)

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

