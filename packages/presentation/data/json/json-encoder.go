package json

import (
	"io"
	"sentinel/packages/common/logger"
	"strings"

	json "github.com/json-iterator/go"
)

var jsonLogger = logger.NewSource("JSON", logger.Default)

type Encoder struct {
    //
}

// Decode given json.
// Returns decoded json and true if there are no errors, false otherwise.
func Decode[T any](input io.Reader) (T, error) {
	var result T

	if err := json.NewDecoder(input).Decode(&result); err != nil {
        jsonLogger.Error("Decoding failed", err.Error(), nil)

		return result, err
	}

	return result, nil
}

func DecodeString[T any](input string) (T, error) {
	return Decode[T](strings.NewReader(input))
}

// Returns `target` argument in json format ([]byte), and true if no errors occured, false otherwise.
// Encode passed `target` argument into json.
func Encode(target any) ([]byte, error) {
	result, err := json.Marshal(target)

	if err != nil {
        jsonLogger.Error("Encoding failed", err.Error(), nil)

		return nil, err
	}

	return result, nil
}

