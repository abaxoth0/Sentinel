package middleware

import (
	"errors"
	"sentinel/packages/presentation/api/http/request"

	"github.com/labstack/echo/v4"
)

type EndpointSensivity int

const (
	InsignificantEndpoint EndpointSensivity = iota
	DefaultEndpoint
	SensitiveEndpoint
)

var sensivityMap = map[EndpointSensivity]bool{
	InsignificantEndpoint: true,
	DefaultEndpoint: true,
	SensitiveEndpoint: true,
}

func (s EndpointSensivity) Validate() error {
	if _, ok := sensivityMap[s]; !ok {
		return errors.New("Sensivity with doesn't exists")
	}
	return nil
}

const sensivityKey = "endpoint_sensivity"

func Sensivity(s EndpointSensivity) echo.MiddlewareFunc {
	if err := s.Validate(); err != nil {
		log.Panic("Failed to set endpoint sensivity", err.Error(), nil)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(sensivityKey, s)
			return next(ctx)
		}
	}
}

// Can be used only after Sensivity middleware beign applied, otherwise will cause panic.
func GetSensivity(ctx echo.Context) EndpointSensivity {
	s, ok := ctx.Get(sensivityKey).(EndpointSensivity)
	if !ok {
		log.Panic(
			"Failed to get endpoint sensivity",
			"Sensivity wasn't found in request context, check if Sensivity middleware applied correctly",
			request.GetMetadata(ctx),
		)
		return -1
	}
	return s
}

