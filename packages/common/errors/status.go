package errs

import "net/http"

type Status struct {
	Status  int
	Message string
}

func (e *Status) Error() string {
	return e.Message
}

func NewStatusError(message string, status int) *Status {
    return &Status{status, message}
}

func IsStatusError(err error) (bool, *Status) {
	e, is := err.(*Status)

	return is, e
}

var StatusInternalError = NewStatusError(
    "Internal Server Error",
    http.StatusInternalServerError,
)

var StatusNotFound = NewStatusError(
    "Запрошенный ресурс не был найден",
    http.StatusNotFound,
)

var StatusTimeout = NewStatusError(
    "Превышено время ожидания",
    http.StatusRequestTimeout,
)

