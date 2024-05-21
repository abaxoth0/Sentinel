package json

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sentinel/packages/net"
)

// Decode given request body.
// Returns decoded json and true if there are no errors, false otherwise.
func Decode[T interface{} | net.AuthRequestBody](body io.ReadCloser, w http.ResponseWriter) (T, bool) {
	var result T

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		log.Printf("\n[ ERROR ] Failed to decode JSON\n%s", err)

		return result, false
	}

	return result, true
}

// Encode passed `target` argument into json.
// Returns `target` argument in json format (type []byte), and true if no errors occured, false otherwise.
func Encode(target interface{}, w http.ResponseWriter) ([]byte, bool) {
	result, err := json.Marshal(target)

	if err != nil {
		log.Printf("\n[ ERROR ] Failed to marshal json\n%s", err)

		return nil, false
	}

	return result, true
}
