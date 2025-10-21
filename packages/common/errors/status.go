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

const (
	Desync int 		= 490
	SessionRevoked  = 491
)

var customStatusesTexts map[int]string = map[int]string{
	Desync: "Data Desynchronization",
	SessionRevoked: "Your session was revoked",
}

// Do the same as http.StatusText(), but also supports custom status codes
func StatusText(status int) string {
	text := http.StatusText(status)
	if text == "" {
		var ok bool
		text, ok = customStatusesTexts[status]
		if !ok {
			return "Unknown Error"
		}
	}
	return text
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
    panic(fmt.Sprintf("Error status range must be between 100 and 599, but got - %d", e.status))
}

// Creates new status error.
// Status must be between 100 and 599 - any other value will cause panic.
func NewStatusError(message string, status int) *Status {
    if status < 100 || status > 599 {
        panic(fmt.Sprintf("Error status range must be between 100 and 599, but got - %d", status))
    }
    return &Status{status, message}
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

var StatusUnauthorized = NewStatusError(
    "Вы не авторизованы",
    http.StatusUnauthorized,
)

var StatusSessionRevoked = NewStatusError(
	customStatusesTexts[SessionRevoked],
	SessionRevoked,
)

