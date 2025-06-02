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
	reqInfo := controller.RequestInfo(ctx)

	controller.Logger.Info("Activating user..." + reqInfo)

    token := ctx.Param("token")

    if strings.ReplaceAll(token, " ", "") == "" {
		controller.Logger.Error("Failed to activate user" + reqInfo, tokenIsMissing.Error())
        return tokenIsMissing
    }

    if err := DB.Database.Activate(token); err != nil {
		controller.Logger.Error("Failed to activate user" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info("Activating user: OK" + reqInfo)

    return ctx.NoContent(http.StatusOK)
}

func Resend(ctx echo.Context) error {
	reqInfo := controller.RequestInfo(ctx)

	controller.Logger.Info("Resending activation email..." + reqInfo)

    var body datamodel.LoginBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
		controller.Logger.Error("Failed to resend activation email" + reqInfo, err.Error())
        return err
    }

    user, err := DB.Database.FindUserByLogin(body.Login)
    if err != nil {
		controller.Logger.Error("Failed to resend activation email" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if user.IsActive() {
		message := "User already active"

		controller.Logger.Error("Failed to resend activation email" + reqInfo, message)

        return echo.NewHTTPError(
            http.StatusConflict,
            message,
        )
    }

	controller.Logger.Trace("Creating activation token..." + reqInfo)

    tk, err := token.NewActivationToken(user.ID, user.Login, user.Roles)
    if err != nil {
		controller.Logger.Error("Failed to create activation token" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Trace("Creating activation token..." + reqInfo)
	controller.Logger.Trace("Creating and enqueueing activation email" + reqInfo)

    email.CreateAndEnqueueActivationEmail(user.Login, tk.String())
    if err != nil {
		controller.Logger.Error("Failed to create and enqueue activation email" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Trace("Creating and enqueueing activation email: OK" + reqInfo)
	controller.Logger.Info("Resending activation email: OK" + reqInfo)

    return ctx.NoContent(http.StatusOK)
}

