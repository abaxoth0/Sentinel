package cachecontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	ActionMapper "sentinel/packages/infrastructure/mappers/action"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"

	"github.com/labstack/echo/v4"
)

// @Summary 		Flush cache
// @Description 	Delete all cache. Only users with "admin" role can do that, even if they have enough permission to do that
// @ID 				drop-cache
// @Tags			cache
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/cache [delete]
// @Security		BearerAuth
func Drop(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Crealing cache...", reqMeta)

    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
		controller.Logger.Error("Failed to clear cache", err.Error(), reqMeta)
        return controller.HandleTokenError(ctx, err)
    }

    act := ActionMapper.BasicActionDTOFromClaims(accessToken.Claims.(*token.Claims))

	if err := authz.User.DropCache(act.RequesterRoles); err != nil {
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

