package docscontroller

import (
	"sentinel/packages/common/config"
	"sentinel/packages/infrastructure/auth/authz"
	controller "sentinel/packages/presentation/api/http/controllers"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @Summary 		This page
// @Description 	API Documentation
// @ID 				api-docs
// @Security		CSRF_Header
// @Security		CSRF_Cookie
func Swagger(ctx echo.Context) error {
	if !config.Debug.Enabled {
		payload := controller.GetAccessTokenPayload(ctx)

		if err := authz.User.AccessAPIDocs(payload.Roles); err != nil {
			return controller.ConvertErrorStatusToHTTP(err)
		}
	}

	return echoSwagger.WrapHandler(ctx)
}
