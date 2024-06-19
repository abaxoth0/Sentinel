package json

import (
	"encoding/json"
	"io"
	"log"
	"sentinel/packages/net"
)

// Decode given request body.
// Returns decoded json and true if there are no errors, false otherwise.
func Decode[T map[string]any | net.AuthRequestBody](body io.ReadCloser) (T, bool) {
	var result T

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		log.Printf("\n[ ERROR ] Failed to decode JSON\n%s", err)

		return result, false
	}

	return result, true
}

// Encode passed `target` argument into json.
// Returns `target` argument in json format (type []byte), and true if no errors occured, false otherwise.
func Encode(target interface{}) ([]byte, bool) {
	result, err := json.Marshal(target)

	if err != nil {
		log.Printf("\n[ ERROR ] Failed to marshal json\n%s", err)

		return nil, false
	}

	return result, true
}
