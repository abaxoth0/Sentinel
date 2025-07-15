package postgres

import (
	"net/http"
	Error "sentinel/packages/common/errors"
)

var loginAlreadyInUse = Error.NewStatusError(
    "Login already in use",
    http.StatusConflict,
)

var invalidActivationTokenFormat = Error.NewStatusError(
    "invalid activation token format. (UUID expected)",
    http.StatusUnprocessableEntity,
)

var activationTokenExpired = Error.NewStatusError(
    "Токен активации истёк. Запросите повторную активацию аккаунта.",
    http.StatusGone,
)

var activationNotFound = Error.NewStatusError(
    "Activation token wasn't found",
    http.StatusNotFound,
)

