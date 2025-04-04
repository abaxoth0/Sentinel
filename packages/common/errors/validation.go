package errs

type Validation struct {
	message string
}

func (e *Validation) Error() string {
	return e.message
}

func NewValidationError(message string) *Validation {
    return &Validation{message}
}

var NoValue = NewValidationError("validation error: no value")
var InvalidValue = NewValidationError("validation error: invalid value")

