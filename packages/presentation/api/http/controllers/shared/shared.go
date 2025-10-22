// This module contains logic that can be shared across other controllers.
// Don't put helpers function here, use controller module instead - this module contains actual logic,
// Like: create CSRF token, create or update session, update location et cetera.
// Also this module doesn't contain any endpoints handlers.
package sharedcontroller

import (
	"crypto/rand"
	"encoding/base64"
	Error "sentinel/packages/common/errors"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"

	"github.com/labstack/echo/v4"
)

func NewCSRFToken(ctx echo.Context) (string, *Error.Status) {
	reqMeta := request.GetMetadata(ctx)

	controller.Log.Trace("Generating CSRF token...", reqMeta)

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		controller.Log.Error("Failed to generate CSRF token", err.Error(), reqMeta)
		return "", Error.StatusInternalError
	}
	tokenStr := base64.RawURLEncoding.EncodeToString(token)

	controller.Log.Trace("Generating CSRF token: OK", reqMeta)

	return tokenStr, nil
}
