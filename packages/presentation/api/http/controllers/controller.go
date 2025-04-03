package controller

import (
	"net/http"
	datamodel "sentinel/packages/presentation/data"

	"github.com/labstack/echo/v4"
)

func BindAndValidate[T datamodel.RequestValidator](ctx echo.Context, dest T) error {
    if err := ctx.Bind(&dest); err != nil {
        return err
    }

    if err := dest.Validate(); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    return nil
}

