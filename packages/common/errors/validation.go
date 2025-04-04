package errs

type Validation struct {
	Message string
}

func (e *Validation) Error() string {
	return e.Message
}

func NewValidationError(message string) *Validation {
    return &Validation{message}
}

var NoValue = NewValidationError("validation error: no value")
var InvalidValue = NewValidationError("validation error: invalid value")
