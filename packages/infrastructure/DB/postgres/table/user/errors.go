package usertable

import (
	"net/http"
	Error "sentinel/packages/common/errors"
)

var loginAlreadyInUse = Error.NewStatusError(
    "Login already in use",
    http.StatusConflict,
)

