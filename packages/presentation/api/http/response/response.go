package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var Unauthorized = echo.NewHTTPError(
    http.StatusUnauthorized,
    "Вы не автозированы",
)

var FailedToDecodeRequestBody = echo.NewHTTPError(
    http.StatusBadRequest,
    "Failed to decode request body",
)

