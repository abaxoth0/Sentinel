package json

import (
	"io"
	"sentinel/packages/common/encoding"

	json "github.com/json-iterator/go"
)

type Encoder struct {
    //
}

// Decode given json.
// Returns decoded json and true if there are no errors, false otherwise.
func Decode[T any](input io.Reader) (T, error) {
	var result T

	if err := json.NewDecoder(input).Decode(&result); err != nil {
        encoding.Log.Error("Failed to decode JSON", err.Error(), nil)

		return result, err
	}

	return result, nil
}

