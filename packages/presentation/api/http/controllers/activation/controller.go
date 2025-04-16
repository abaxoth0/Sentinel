package activationcontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/email"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	datamodel "sentinel/packages/presentation/data"
	"strings"

	"github.com/labstack/echo/v4"
)

var tokenIsMissing = echo.NewHTTPError(
    http.StatusBadRequest,
    "Token is missing",
)

func Activate(ctx echo.Context) error {
    token := ctx.Param("token")

    if strings.ReplaceAll(token, " ", "") == "" {
        return tokenIsMissing
    }

    if err := DB.Database.Activate(token); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    return ctx.NoContent(http.StatusOK)
}

func Resend(ctx echo.Context) error {
    var body datamodel.LoginBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    user, err := DB.Database.FindUserByLogin(body.Login)
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if user.IsActive {
        return echo.NewHTTPError(
            http.StatusConflict,
            "User already active",
        )
    }

    tk, err := token.NewActivationToken(user.ID, user.Login, user.Roles)
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    email.CreateAndEnqueueActivationEmail(user.Login, tk.String())
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    return ctx.NoContent(http.StatusOK)
}

