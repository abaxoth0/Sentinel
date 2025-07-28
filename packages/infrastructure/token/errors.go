package token

import (
	"net/http"
	Error "sentinel/packages/common/errors"
)

// TODO handle all this errors

var TokenMalformed = Error.NewStatusError(
    "Token is malformed or has invalid format",
    // According to RFC 7235 (https://datatracker.ietf.org/doc/html/rfc7235#section-3.1)
    // 401 response status code indicates that the request lacks VALID authentication credentials,
    // no matter if token was invalid, missing or auth creditinals is invalid.
    http.StatusUnauthorized,
)

var TokenExpired = Error.NewStatusError(
    "Token expired",
    http.StatusUnauthorized,
)

var InvalidToken = Error.NewStatusError(
    "Invalid Token",
    http.StatusBadRequest,
)

var TokenInvalidSignature = Error.NewStatusError(
    "Invalid Token Signature",
    http.StatusBadRequest,
)

var TokenModified = Error.NewStatusError(
    "Invalid Token (and you know that)",
    http.StatusBadRequest,
)

var TokenMissingRequiredClaims = Error.NewStatusError(
    "At least one of required token claims is missing",
    http.StatusBadRequest,
)

var TokenAudienceDoesNotExists = Error.NewStatusError(
	"Audience doesn't exists",
	http.StatusBadRequest,
)

var TokenAudienceIsNotSpecified = Error.NewStatusError(
	"Audience isn't specified",
	http.StatusBadRequest,
)

var TokenAudienceMismatch = Error.NewStatusError(
	"Token not valid for this audience",
	http.StatusBadRequest,
)

func IsTokenError(err *Error.Status) bool {
    return err == TokenMalformed ||
        	err == TokenExpired ||
    		err == TokenInvalidSignature ||
			err == InvalidToken ||
			err == TokenMissingRequiredClaims ||
			err == TokenAudienceDoesNotExists ||
			err == TokenAudienceMismatch ||
			err == TokenAudienceIsNotSpecified
}

