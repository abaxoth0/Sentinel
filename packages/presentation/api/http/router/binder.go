package router

import (
	"io"
	"sentinel/packages/presentation/api/http/response"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

type binder struct {
    //
}

func (b *binder) Bind(i interface{}, ctx echo.Context) error {
    body, err := io.ReadAll(ctx.Request().Body)

    if err != nil {
        return response.FailedToReadRequestBody
    }

    if err := jsoniter.Unmarshal(body, i); err != nil {
        return response.FailedToDecodeRequestBody
    }

    return nil
}

