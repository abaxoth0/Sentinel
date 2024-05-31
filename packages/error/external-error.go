// custom error
package externalerror

// This error's text will be send to user in response.
type Error struct {
	Message string
	Status  int
}

func (e *Error) Error() string {
	return e.Message
}

func New(message string, status int) *Error {
	return &Error{message, status}
}

// Check is error type - ExternalError
func Is(err error) (bool, *Error) {
	e, is := err.(*Error)

	return is, e
}
