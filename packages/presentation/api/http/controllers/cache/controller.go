package cachecontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func Drop(ctx echo.Context) error {
    accessToken, err := token.GetAccessToken(ctx.Request().Header.Get("Authorization"))

    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    filter, err := UserMapper.FilterDTOFromClaims(UserMapper.NoTarget, accessToken.Claims.(jwt.MapClaims))

    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if err := authorization.Authorize(
        authorization.Action.Drop,
        authorization.Resource.Cache,
        filter.RequesterRoles,
    ); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if err := cache.Client.FlushAll(); err != nil {
        return err
    }

    return ctx.NoContent(http.StatusOK)
}

