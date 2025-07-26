package docscontroller

import (
	"sentinel/packages/common/config"
	"sentinel/packages/infrastructure/auth/authz"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Swagger(ctx echo.Context) error {
	if !config.Debug.Enabled {
		accessToken, err := controller.GetAccessToken(ctx)
		if err != nil {
			return controller.ConvertErrorStatusToHTTP(err)
		}

		payload := UserMapper.PayloadFromClaims(accessToken.Claims.(*token.Claims))

		if err := authz.User.AccessAPIDocs(payload.Roles); err != nil {
			return controller.ConvertErrorStatusToHTTP(err)
		}
	}

	return echoSwagger.WrapHandler(ctx)
}
