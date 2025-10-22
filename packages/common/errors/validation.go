package errs

import (
	"net/http"
)

// Do not create new instances of this error,
// instead use NoValue and InvalidValue sentinel errors.
type Validation struct {
	message string
}

func (e *Validation) Error() string {
	return e.message
}

// Converts Validation error to Status error.
// Returned error by default will have "Bad Request" status.
// Will return Error.StatusInternalError if for some reason
// error is not NoValue or InvalidValue.
func (e *Validation) ToStatus(noValueMsg string, invalidValueMsg string) *Status {
	if e == NoValue {
		return NewStatusError(noValueMsg, http.StatusBadRequest)
	}
	if e == InvalidValue {
		return NewStatusError(invalidValueMsg, http.StatusBadRequest)
	}
	panic("Invalid validation error: Expected NoValue or InvalidValue")
}

func NewValidationError(message string) *Validation {
	return &Validation{message}
}

var NoValue = NewValidationError("validation error: no value")
var InvalidValue = NewValidationError("validation error: invalid value")
