package postgres

import (
	"net/http"
	Error "sentinel/packages/common/errors"
)

var loginAlreadyInUse = Error.NewStatusError(
    "Login already in use",
    http.StatusConflict,
)

var userNotFound = Error.NewStatusError(
    "Пользователь не был найден",
    http.StatusNotFound,
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

// Returns userNotFound if 'err' is Error.StatusNotFound,
// otherwise returns 'err'
func tryMapToUserNotFound(err *Error.Status) *Error.Status {
    if err == Error.StatusNotFound {
        return userNotFound
    }
    return err
}

