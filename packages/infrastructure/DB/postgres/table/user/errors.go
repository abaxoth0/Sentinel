package usertable

import (
	"net/http"
	Error "sentinel/packages/common/errors"
)

var loginAlreadyInUse = Error.NewStatusError(
	"This login is already in use by another user",
	http.StatusConflict,
)
