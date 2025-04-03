package activationcontroller

import (
	"net/http"
	"sentinel/packages/infrastructure/DB"
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
        return echo.NewHTTPError(err.Status, err.Message)
    }

    return ctx.NoContent(http.StatusOK)
}

func Reactivate(ctx echo.Context) error {
    var body datamodel.LoginBody

    if err := ctx.Bind(&body); err != nil {
        return err
    }

    if err := body.Validate(); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    if err := DB.Database.Reactivate(body.Login); err != nil {
        return echo.NewHTTPError(err.Status, err.Message)
    }

    return ctx.NoContent(http.StatusOK)
}

