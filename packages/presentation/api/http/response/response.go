package response

import (
	"net/http"

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

