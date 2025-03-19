package response

import (
	"net/http"
	datamodel "sentinel/packages/presentation/data"

	"github.com/labstack/echo/v4"
)


var Unauthorized = echo.NewHTTPError(
    http.StatusUnauthorized,
    "Вы не автозированы",
)

var FailedToReadRequestBody = echo.NewHTTPError(
    http.StatusBadRequest,
    "Failed to read request body",
)

var FailedToDecodeRequestBody = echo.NewHTTPError(
    http.StatusBadRequest,
    "Failed to decode request body",
)

var RequestMissingUid = echo.NewHTTPError(
    http.StatusBadRequest,
    datamodel.InvalidUID.Error(),
)

var RequestMissingLogin = echo.NewHTTPError(
    http.StatusBadRequest,
    datamodel.InvalidLogin.Error(),
)

var RequestMissingPassword = echo.NewHTTPError(
    http.StatusBadRequest,
    datamodel.InvalidPassword.Error(),
)

var RequestMissingRoles = echo.NewHTTPError(
    http.StatusBadRequest,
    datamodel.InvalidRoles.Error(),
)

