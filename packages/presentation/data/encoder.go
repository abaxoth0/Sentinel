package datamodel

import (
	"io"
	"sentinel/packages/common/encoding/json"
)

func Encode(data interface{}) ([]byte, error) {
    return json.Encode(data)
}

func Decode[T interface{}](input io.Reader) (T, error) {
    return json.Decode[T](input)
}

func DecodeString[T interface{}](input string) (T, error) {
    return json.DecodeString[T](input)
}

