package activationcontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/email"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	RequestBody "sentinel/packages/presentation/data/request"
	"strings"

	"github.com/labstack/echo/v4"
)

var tokenIsMissing = echo.NewHTTPError(
	http.StatusBadRequest,
	"Token is missing",
)

// @Summary 		Activate user
// @Description 	Activate user
// @ID 				activate
// @Tags			activation
// @Param 			token path string true "Activation token"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,500	{object} responsebody.Error
// @Router			/v1/user/activate/{token} [get]
func Activate(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	token := ctx.Param("token")

	if strings.ReplaceAll(token, " ", "") == "" {
		controller.Log.Error("Failed to activate user", tokenIsMissing.Error(), reqMeta)
		return tokenIsMissing
	}

	if err := DB.Database.Activate(token); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

// @Summary 		Resend activation token
// @Description 	Create and send new activation token to user
// @ID 				resend-activation-token
// @Tags			activation
// @Param 			login body requestbody.UserLogin true "Login of not activated user to whom token should be sent"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,500	{object} responsebody.Error
// @Router			/v1/user/activate/resend [put]
func Resend(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	controller.Log.Info("Resending activation email...", reqMeta)

	var body RequestBody.UserLogin

	if e := controller.BindAndValidate(ctx, &body); e != nil {
		return e
	}

	user, err := DB.Database.GetUserByLogin(body.Login)
	if err != nil {
		return err
	}

	if user.IsActive() {
		errMsg := "User already active"
		controller.Log.Error("Failed to resend activation email", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusConflict, errMsg)
	}

	tk, err := token.NewActivationToken(user.ID, user.Login)
	if err != nil {
		return err
	}

	err = email.EnqueueEmail(email.ActivationEmail, user.Login, email.Substitutions{
		email.TokenPlaceholder: tk.String(),
	})
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}
