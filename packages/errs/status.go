package errs

type Status struct {
	Message string
	Status  int
}

func (e *Status) Error() string {
	return e.Message
}

func NewStatusError(message string, status int) *Status {
	return &Status{message, status}
}

func IsStatusError(err error) (bool, *Status) {
	e, is := err.(*Status)

	return is, e
}

