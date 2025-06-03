package activationcontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/email"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	datamodel "sentinel/packages/presentation/data"
	"strings"

	"github.com/labstack/echo/v4"
)

var tokenIsMissing = echo.NewHTTPError(
    http.StatusBadRequest,
    "Token is missing",
)

func Activate(ctx echo.Context) error {
	reqMeta, err := request.GetLogMeta(ctx)
	if err != nil {
		controller.Logger.Panic("Failed to get log meta for the request",err.Error(), nil)
		return err
	}

	controller.Logger.Info("Activating user...", nil)

    token := ctx.Param("token")

    if strings.ReplaceAll(token, " ", "") == "" {
		controller.Logger.Error("Failed to activate user", tokenIsMissing.Error(), reqMeta)
        return tokenIsMissing
    }

    if err := DB.Database.Activate(token); err != nil {
		controller.Logger.Error("Failed to activate user", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info("Activating user: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

func Resend(ctx echo.Context) error {
	reqMeta, err := request.GetLogMeta(ctx)
	if err != nil {
		controller.Logger.Panic("Failed to get log meta for the request",err.Error(), nil)
		return err
	}

	controller.Logger.Info("Resending activation email...", reqMeta)

    var body datamodel.LoginBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
		controller.Logger.Error("Failed to resend activation email", err.Error(), reqMeta)
        return err
    }

    user, e := DB.Database.FindUserByLogin(body.Login)
    if e != nil {
		controller.Logger.Error("Failed to resend activation email", e.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(e)
    }

    if user.IsActive() {
		message := "User already active"

		controller.Logger.Error("Failed to resend activation email", message, reqMeta)

        return echo.NewHTTPError(
            http.StatusConflict,
            message,
        )
    }

	controller.Logger.Trace("Creating activation token...", reqMeta)

    tk, e := token.NewActivationToken(user.ID, user.Login, user.Roles)
    if e != nil {
		controller.Logger.Error("Failed to create activation token", e.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(e)
    }

	controller.Logger.Trace("Creating activation token...", reqMeta)
	controller.Logger.Trace("Creating and enqueueing activation email", reqMeta)

	e = email.CreateAndEnqueueActivationEmail(user.Login, tk.String())
    if err != nil {
		controller.Logger.Error("Failed to create and enqueue activation email", e.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(e)
    }

	controller.Logger.Trace("Creating and enqueueing activation email: OK", reqMeta)
	controller.Logger.Info("Resending activation email: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

