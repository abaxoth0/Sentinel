package errs

import (
	"fmt"
	"net/http"
)

type Status struct {
	status  int
	message string
}

func (e *Status) Error() string {
	return e.message
}

func (e *Status) Status() int {
    return e.status
}

type errorSide string

const (
    ClientSide errorSide = "client"
    ServerSide errorSide = "server"
)

// Side returns whether the status represents a client or server error.
//
// Returns ClientSide for status codes 400-499.
//
// Returns ServerSide for status codes 500-599.
//
// Panics if the status isn't in either of these ranges.
func (e *Status) Side() errorSide {
    if e.status > 399 && e.status < 500 {
        return ClientSide
    }
    if e.status > 500 && e.status < 600{
        return ServerSide
    }
    panic(fmt.Sprintf("invalid error status range: must be between 100 and 599, but got - %d", e.status))
}

// Creates new status error.
// Status must be between 100 and 599 - any other value will cause panic.
func NewStatusError(message string, status int) *Status {
    if status < 100 || status > 599 {
        panic(fmt.Sprintf("invalid error status range: must be between 100 and 599, but got - %d", status))
    }
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

