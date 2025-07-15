package rolescontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authz"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	"strings"

	"github.com/labstack/echo/v4"
)

var serviceIdIsNotSpecified = echo.NewHTTPError(
    http.StatusBadRequest,
    "Service ID is not specified",
)


// @Summary 		Get all service roles
// @Description 	Get list of all roles that exists in the specified service
// @ID 				get-all-roles
// @Tags			roles
// @Param 			serviceID path string true "ID of the service which roles you want to get"
// @Accept			json
// @Produce			json
// @Success			200 			{array} 	string
// @Failure			400,401,500 	{object} 	responsebody.Error
// @Router			/roles/{serviceID} [get]
func GetAll(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

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

    return ctx.JSON(http.StatusOK, roles)
}

