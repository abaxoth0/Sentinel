package json

import (
	"encoding/json"
	"io"
	"log"
	"strings"
)

type Encoder struct {
    //
}

// Decode given json.
// Returns decoded json and true if there are no errors, false otherwise.
func Decode[T interface{}](input io.Reader) (T, error) {
	var result T

	if err := json.NewDecoder(input).Decode(&result); err != nil {
		log.Printf("\n[ ERROR ] Failed to decode JSON\n%s", err)

		return result, err
	}

	return result, nil
}

func DecodeString[T interface{}](input string) (T, error) {
	return Decode[T](strings.NewReader(input))
}

// Returns `target` argument in json format ([]byte), and true if no errors occured, false otherwise.
// Encode passed `target` argument into json.
func Encode(target interface{}) ([]byte, error) {
	result, err := json.Marshal(target)

	if err != nil {
		log.Printf("\n[ ERROR ] Failed to marshal json\n%s", err)

		return nil, err
	}

	return result, nil
}

