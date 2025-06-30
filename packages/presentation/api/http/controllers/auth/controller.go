package authcontroller

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authn"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	datamodel "sentinel/packages/presentation/data"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Login(ctx echo.Context) error {
    var body datamodel.LoginPasswordBody
    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Authenticating user '"+body.Login+"'...", reqMeta)

    user, err := DB.Database.FindUserByLogin(body.Login)
    if err != nil {
		controller.Logger.Error("Failed to authenticate user '"+body.Login+"'...", err.Error(), reqMeta)

        if err.Side() == Error.ClientSide {
            return echo.NewHTTPError(
                authn.InvalidAuthCreditinals.Status(),
                authn.InvalidAuthCreditinals.Error(),
            )
        }
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if err := authn.CompareHashAndPassword(user.Password, body.Password); err != nil {
		controller.Logger.Error("Failed to authenticate user '"+body.Login+"'", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

    payload := &UserDTO.Payload{
        ID: user.ID,
        Login: user.Login,
        Roles: user.Roles,
    }

    accessToken, refreshToken, err := token.NewAuthTokens(payload)
    if err != nil {
		controller.Logger.Error("Failed to authenticate user '"+body.Login+"'", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

    ctx.SetCookie(newAuthCookie(refreshToken))

	controller.Logger.Info("Authenticating user '"+body.Login+"': OK", reqMeta)

    return ctx.JSON(
        http.StatusOK,
        datamodel.TokenResponseBody{
            Message: "Пользователь успешно авторизован",
            AccessToken: accessToken.String(),
            ExpiresIn: int(accessToken.TTL()) / 1000,
        },
    )
}

func Logout(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

    authCookie, err := controller.GetAuthCookie(ctx)
    if err != nil {
        return err
    }

	tk, e := token.ParseSingedToken(authCookie.Value, config.Secret.RefreshTokenPublicKey)
	if e != nil {
		controller.Logger.Error("Failed to parse refresh token", e.Error(), reqMeta)
		return controller.ConvertErrorStatusToHTTP(e)
	}

	uid, er := tk.Claims.GetSubject()
	if er != nil {
		controller.Logger.Error("Invalid refresh token claims", er.Error(), reqMeta)
		return Error.StatusInternalError
	}

	controller.Logger.Info("User '"+uid+"' logging out...", reqMeta)

    controller.DeleteCookie(ctx, authCookie)

	controller.Logger.Info("User '"+uid+"' logging out: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

func Refresh(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Refreshing auth tokens...", reqMeta)

    currentRefreshToken, err := controller.GetRefreshToken(ctx)
    if err != nil {
		controller.Logger.Error("Failed to refresh auth tokens", err.Error(), reqMeta)
        return controller.HandleTokenError(ctx, err)
    }

    payload, err := UserMapper.PayloadFromClaims(currentRefreshToken.Claims.(jwt.MapClaims))
    if err != nil {
		controller.Logger.Error("Failed to refresh auth tokens", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

    accessToken, refreshToken, err := token.NewAuthTokens(payload)
    if err != nil {
		controller.Logger.Error("Failed to refresh auth tokens", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

    ctx.SetCookie(newAuthCookie(refreshToken))

	controller.Logger.Info("Refreshing auth tokens: OK", reqMeta)

    return ctx.JSON(
        http.StatusOK,
        datamodel.TokenResponseBody{
            Message: "Токены успешно обновлены",
            AccessToken: accessToken.String(),
            ExpiresIn: int(accessToken.TTL()) / 1000,
        },
    )
}

func Verify(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Verifying access token...", reqMeta)

	// If token is invalid (expired, malformed etc) then this method will return error
    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
		controller.Logger.Error("Verifying access token: ERROR", err.Error(), reqMeta)
        return controller.HandleTokenError(ctx, err)
    }

    payload, err := UserMapper.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))
    if err != nil {
		controller.Logger.Error("Failed to verify access token", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info("Verifying access token: OK", reqMeta)

    return ctx.JSON(http.StatusOK, payload)
}

