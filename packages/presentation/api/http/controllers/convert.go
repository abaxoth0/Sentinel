package controller

import (
	Error "sentinel/packages/common/errors"

	"github.com/labstack/echo/v4"
)

func ConvertErrorStatusToHTTP(err *Error.Status) *echo.HTTPError {
    return echo.NewHTTPError(err.Status, err.Message)
}

