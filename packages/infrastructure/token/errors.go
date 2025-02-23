package token

import (
	"net/http"
	Error "sentinel/packages/errors"
)


var accessTokenExpired = Error.NewStatusError(
    "Access Token expired",
    http.StatusUnauthorized,
)

var invalidAccessToken =Error.NewStatusError(
    "Invalid Access Token",
    http.StatusBadRequest,
)

var unauthorized =Error.NewStatusError(
    "Вы не авторизованы",
    http.StatusUnauthorized,
)

var invalidRefreshToken =Error.NewStatusError(
    "Invalid Refresh Token",
    http.StatusBadRequest,
)

var refreshTokenExpired =Error.NewStatusError(
    "Refresh Token Expired",
    // Not sure that status 409 is OK for this case,
    // currently this tells user that there are conflict with server and him,
    // and reason of conflict in next: User assumes that he authorized but it's
    // wrong, cuz refresh token expired.
    // More likely will be better to use status 401 (unathorized) in this case,
    // but once againg - i'm not sure.
    http.StatusConflict,
)

