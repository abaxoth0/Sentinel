package cachecontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Drop(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Crealing cache...", reqMeta)

    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
		controller.Logger.Error("Failed to clear cache", err.Error(), reqMeta)
        return controller.HandleTokenError(ctx, err)
    }

    filter, err := UserMapper.BasicActionDTOFromClaims(accessToken.Claims.(jwt.MapClaims))
    if err != nil {
		controller.Logger.Error("Failed to clear cache", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	if err := authz.User.DropCache(filter.RequesterRoles); err != nil {
		controller.Logger.Error("Failed to clear cache", err.Error(), reqMeta)
		return err
	}

    if err := cache.Client.FlushAll(); err != nil {
		controller.Logger.Error("Failed to clear cache", err.Error(), reqMeta)
        return err
    }

	controller.Logger.Info("Crealing cache: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

