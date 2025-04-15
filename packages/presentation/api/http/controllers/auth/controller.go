package authcontroller

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authentication"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	datamodel "sentinel/packages/presentation/data"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func Login(ctx echo.Context) error {
    var body datamodel.LoginPasswordBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    user, err := DB.Database.FindAnyUserByLogin(body.Login)
    if err != nil {
        if err.Side() == Error.ClientSide {
            return echo.NewHTTPError(
                authentication.InvalidAuthCreditinals.Status(),
                authentication.InvalidAuthCreditinals.Error(),
            )
        }
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if err := authentication.CompareHashAndPassword(user.Password, body.Password); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    payload := &UserDTO.Payload{
        ID: user.ID,
        Login: user.Login,
        Roles: user.Roles,
    }

    accessToken, refreshToken, err := token.NewAuthTokens(payload)
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    ctx.SetCookie(newAuthCookie(refreshToken))

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
    authCookie, err := controller.GetAuthCookie(ctx)
    if err != nil {
        return err
    }

    controller.DeleteCookie(ctx, authCookie)

    return ctx.NoContent(http.StatusOK)
}

func Refresh(ctx echo.Context) error {
    currentRefreshToken, e := controller.GetRefreshToken(ctx)
    if e != nil {
        if token.IsTokenError(e) {
            authCookie, err := controller.GetAuthCookie(ctx)
            if err != nil {
                return err
            }
            controller.DeleteCookie(ctx, authCookie)
        }
        return echo.NewHTTPError(
            http.StatusUnauthorized,
            e.Error(),
        )
    }

    payload, e := UserMapper.PayloadFromClaims(currentRefreshToken.Claims.(jwt.MapClaims))
    if e != nil {
        return controller.ConvertErrorStatusToHTTP(e)
    }

    accessToken, refreshToken, e := token.NewAuthTokens(payload)
    if e != nil {
        return controller.ConvertErrorStatusToHTTP(e)
    }


    ctx.SetCookie(newAuthCookie(refreshToken))

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
    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    payload, err := UserMapper.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    return ctx.JSON(http.StatusOK, payload)
}

