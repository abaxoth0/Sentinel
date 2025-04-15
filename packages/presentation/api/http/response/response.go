package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var FailedToReadRequestBody = echo.NewHTTPError(
    http.StatusBadRequest,
    "Failed to read request body",
)

var FailedToDecodeRequestBody = echo.NewHTTPError(
    http.StatusBadRequest,
    "Failed to decode request body",
)

