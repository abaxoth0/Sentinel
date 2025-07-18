package dberrors

import (
	"net/http"
	Error "sentinel/packages/common/errors"
)

var LoginAlreadyInUse = Error.NewStatusError(
    "Login already in use",
    http.StatusConflict,
)

var InvalidActivationTokenFormat = Error.NewStatusError(
    "invalid activation token format. (UUID expected)",
    http.StatusUnprocessableEntity,
)

var ActivationTokenExpired = Error.NewStatusError(
    "Токен активации истёк. Запросите повторную активацию аккаунта.",
    http.StatusGone,
)

var ActivationNotFound = Error.NewStatusError(
    "Activation token wasn't found",
    http.StatusNotFound,
)

