package router

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

type serializer struct{}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Serialize converts the input into JSON using jsoniter
func (serializer) Serialize(c echo.Context, v interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())

	if indent != "" {
		enc.SetIndent("", indent)
	}

	return enc.Encode(v)
}

// Deserialize reads the JSON from the request body and decodes it into the input using jsoniter
func (serializer) Deserialize(c echo.Context, v interface{}) error {
	return json.NewDecoder(c.Request().Body).Decode(v)
}
