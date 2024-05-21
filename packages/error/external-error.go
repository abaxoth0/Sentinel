// custom error
package externalerror

// This error's text will be send to user in response.
type ExternalError struct {
	Message string
	Status  int
}

func (e *ExternalError) Error() string {
	return e.Message
}

func New(message string, status int) *ExternalError {
	return &ExternalError{message, status}
}

// Check is error type - ExternalError
func Is(err error) (bool, *ExternalError) {
	e, is := err.(*ExternalError)

	return is, e
}
