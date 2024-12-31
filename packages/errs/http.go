// custom error
package errs

// This error's text will be send to user in response.
type HTTP struct {
	Message string
	Status  int
}

func (e *HTTP) Error() string {
	return e.Message
}

func NewHTTP(message string, status int) *HTTP {
	return &HTTP{message, status}
}

// Check is error type - ExternalError
func Is(err error) (bool, *HTTP) {
	e, is := err.(*HTTP)

	return is, e
}
