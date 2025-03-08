package authcontroller

import (
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/auth/authentication"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/response"
	datamodel "sentinel/packages/presentation/data"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func Login(ctx echo.Context) error {
    body, err := datamodel.Decode[datamodel.AuthRequestBody](ctx.Request().Body)

    if err != nil {
        return response.FailedToDecodeRequestBody
    }

    user, e := authentication.Login(body.Login, body.Password)

    if e != nil {
        return echo.NewHTTPError(e.Status, e.Message)
    }

    accessToken, refreshToken := token.Generate(&UserDTO.Payload{
        ID: user.ID,
        Login: user.Login,
        Roles: user.Roles,
    })

    ctx.SetCookie(newAuthCookie(refreshToken))

    return ctx.JSON(
        http.StatusOK,
        datamodel.TokenResponseBody{
            Message: "Пользователь успешно авторизован",
            AccessToken: accessToken.Value,
        },
    )
}

func Logout(ctx echo.Context) error {
    authCookie, err := getAuthCookie(ctx)

    if err != nil {
        return err
    }

    deleteCookie(ctx, authCookie)

    return ctx.NoContent(http.StatusOK)
}

func Refresh(ctx echo.Context) error {
    authCookie, err := getAuthCookie(ctx)

    if err != nil {
        return err
    }

    oldRefreshToken, e := token.GetRefreshToken(authCookie)

    // if refresh token is either invalid or expired
    if e != nil {
        deleteCookie(ctx, authCookie)

        return echo.NewHTTPError(e.Status, e.Message)
    }

    payload, e := UserMapper.PayloadFromClaims(oldRefreshToken.Claims.(jwt.MapClaims))

    if e != nil {
        return echo.NewHTTPError(e.Status, e.Message)
    }

    accessToken, refreshToken := token.Generate(payload)

    ctx.SetCookie(newAuthCookie(refreshToken))

    return ctx.JSON(
        http.StatusOK,
        datamodel.TokenResponseBody{
            Message: "Токены успешно обновлены",
            AccessToken: accessToken.Value,
        },
    )
}

func Verify(ctx echo.Context) error {
    authHeader := ctx.Request().Header.Get("Authorization")

    accessToken, err := token.GetAccessToken(authHeader)

    if err != nil {
        return echo.NewHTTPError(err.Status, err.Message)
    }

    payload, err := UserMapper.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))

    if err != nil {
        return echo.NewHTTPError(err.Status, err.Message)
    }

    return ctx.JSON(http.StatusOK, payload)
}

