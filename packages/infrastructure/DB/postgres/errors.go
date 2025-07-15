package postgres

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/core/user"
	"strings"
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

func newUsersStateConflictError(newState user.State, ids []string) *Error.Status {
	var message string
	if newState == user.DeletedState {
		message = "Can't delete already deleted user(-s): " + strings.Join(ids, ", ")
	} else {
		message = "Can't restore non-deleted user(-s): " + strings.Join(ids, ", ")
	}
	return Error.NewStatusError(message, http.StatusConflict)
}

