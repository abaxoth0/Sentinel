package net

import (
	"errors"
	"log"
	"net/http"
	"sentinel/packages/models/token"
	"strings"
)

type request struct {
	//
}

var Request = request{}

// Return true if method supported, false otherwise
// (Except OPTIONS, for this method always will be returned false).
// If method isn't supported sends error response (status 405).
// If method OPTIONS then response will not be send.
func (r *request) Preprocessing(w http.ResponseWriter, req *http.Request, method string) bool {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", method) // "POST, GET, OPTIONS, PUT, DELETE"
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Date, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	log.Printf("[ %s ] %s %s", req.RemoteAddr, req.Method, req.RequestURI)

	if req.Method == "OPTIONS" {
		return false
	}

	if req.Method != method {
		Response.Message("Method Not Allowed. Allowed methods: "+method, http.StatusMethodNotAllowed, w)

		r.Print("[ ERROR ] Method allowed", req)

		return false
	}

	return true
}

// Returns access and refresh tokens strings. (In same order as here)
func (r *request) ExtractRawTokens(req *http.Request) (string, string) {
	atk, rtk := "", ""

	authHeaderValue := req.Header.Get("Authorization")

	if authHeaderValue != "" {
		atk = strings.Split(authHeaderValue, "Bearer ")[1]
	}

	authCookie, err := req.Cookie(token.RefreshTokenKey)

	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			panic(err)
		}

		r.Print("[ ERROR ] Auth cookie wasn't found", req)

		return atk, rtk
	}

	rtk = authCookie.Value

	return atk, rtk
}

// Works like "log.Print()", but also attaches some request data to output.
// (IP of device, that sent request; HTTP Method; Requested URI)
//
// Example with `message` = "Authentication successful, user id: ...:
//
// "2020/01/01 12:00:00 [ 127.0.0.1:5000 ] POST /login | Authentication successful, user id: ..."
func (r *request) Print(message string, req *http.Request) {
	log.Printf("[ %s ] %s %s | %s", req.RemoteAddr, req.Method, req.RequestURI, message)
}

// Works like "log.Print()", but also attaches some request data to output.
// (IP of device, that sent request; HTTP Method; Requested URI)
//
// Example with `message` = "Access token expired":
//
// "2020/01/01 12:00:00 [ 127.0.0.1:50000 ] Error: GET /verification | Access token expired"
func (r *request) PrintError(message string, status int, req *http.Request) {
	log.Printf("[ %s ] Error: %s %s | %s", req.RemoteAddr, req.Method, req.RequestURI, message)
}
