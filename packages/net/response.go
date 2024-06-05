package net

import (
	"encoding/json"
	"log"
	"net/http"
)

type response struct {
	//
}

var Response = response{}

// Sends response with status 200 and JSON {"message": "OK"}
// Also handles all possible errors, if one was return, that mean response has already been sent.
func (res *response) OK(w http.ResponseWriter) error {
	body, err := json.Marshal(MessageResponseBody{Message: "OK"})

	if err != nil {
		log.Println("[ ERROR ] Failed to marshal json.")

		e := res.InternalServerError(w)

		if e != nil {
			panic(e)
		}

		return err
	}

	w.WriteHeader(http.StatusOK)

	if err := res.writeBody(body, w); err != nil {
		panic(err)
	}

	return nil
}

// Sends response with passed status and JSON {"message": <message>}
// Also handles all possible errors, if one was return, that mean response has already been sent.
func (res *response) Message(message string, status int, w http.ResponseWriter) error {
	body, err := json.Marshal(MessageResponseBody{Message: message})

	if err != nil {
		e := res.InternalServerError(w)

		if e != nil {
			panic(e)
		}

		return err
	}

	w.WriteHeader(status)

	if err := res.writeBody(body, w); err != nil {
		panic(err)
	}

	return nil
}

// Sends response with status 200 and given `body`.
// If error was return, that mean response has already been sent.
func (res *response) Send(body []byte, w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)

	return res.writeBody(body, w)
}

// Sends response with given message and status, also log request info in terminal.
//
// Returns error if failed to send response (also does log in this case), nil otherwise.
func (res *response) SendError(message string, status int, req *http.Request, w http.ResponseWriter) error {
	err := Response.Message(message, status, w)

	if err != nil {
		Request.PrintError("Failed to send response", 500, req)

		return err
	}

	Request.PrintError(message, status, req)

	return nil
}

func (res *response) InternalServerError(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusInternalServerError)

	body, err := json.Marshal(MessageResponseBody{Message: "Internal Server Error"})

	if err != nil {
		log.Println("[ ERROR ] Failed to marshal json (status 500)")

		return err
	}

	// IMPORTANT
	// Don't use `writeBody` here, it may cause infinite recursion.
	// (cuz this method used inside `writeBody`)
	if _, writeError := w.Write(body); writeError != nil {
		log.Println("[ ERROR ] Failed to write in response body (status 500)")

		panic(writeError)
	}

	return nil
}

// Writes given `body` in response. Return nil if success, error otherwise.
// Also in error case sends error response with status 500. (Using `InternalServerError` method)
func (res *response) writeBody(body []byte, w http.ResponseWriter) error {
	if _, writeError := w.Write(body); writeError != nil {
		log.Println("[ ERROR ] Failed to write in response body (status 500)")

		err := res.InternalServerError(w)

		if err != nil {
			panic(err)
		}

		return writeError
	}

	return nil
}
