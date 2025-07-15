package docscontroller

import (
	"sentinel/packages/common/config"
	"sentinel/packages/infrastructure/auth/authz"
	usermapper "sentinel/packages/infrastructure/mappers/user"
	controller "sentinel/packages/presentation/api/http/controllers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Swagger(ctx echo.Context) error {
	if !config.Debug.Enabled {
		accessToken, err := controller.GetAccessToken(ctx)
		if err != nil {
			return controller.ConvertErrorStatusToHTTP(err)
		}

		payload, err := usermapper.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))
		if err != nil {
			return controller.ConvertErrorStatusToHTTP(err)
		}

		if err := authz.User.AccessAPIDocs(payload.Roles); err != nil {
			return controller.ConvertErrorStatusToHTTP(err)
		}
	}

	return echoSwagger.WrapHandler(ctx)
}
