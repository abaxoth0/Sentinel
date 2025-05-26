package cachecontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	controller "sentinel/packages/presentation/api/http/controllers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Drop(ctx echo.Context) error {
	reqInfo := controller.RequestInfo(ctx)

	controller.Logger.Info("Crealing cache..." + reqInfo)

    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
		controller.Logger.Error("Failed to clear cache" + reqInfo, err.Error())
        return controller.HandleTokenError(ctx, err)
    }

    filter, err := UserMapper.BasicActionDTOFromClaims(accessToken.Claims.(jwt.MapClaims))
    if err != nil {
		controller.Logger.Error("Failed to clear cache" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if err := authz.Authorize(
        authz.Action.Drop,
        authz.Resource.Cache,
        filter.RequesterRoles,
    ); err != nil {
		controller.Logger.Error("Failed to clear cache" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(err)
    }
    if err := cache.Client.FlushAll(); err != nil {
		controller.Logger.Error("Failed to clear cache" + reqInfo, err.Error())
        return err
    }

	controller.Logger.Info("Crealing cache: OK" + reqInfo)

    return ctx.NoContent(http.StatusOK)
}

