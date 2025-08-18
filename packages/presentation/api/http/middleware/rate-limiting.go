package middleware

import (
	"net/http"
	"sentinel/packages/presentation/api/http/request"
	ResponseBody "sentinel/packages/presentation/data/response"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type rateLimiter struct {
	//
}

func rateLimiterIdentifierExtractor(ctx echo.Context) (string, error) {
	return ctx.RealIP(), nil
}

func rateLimiterDenyHandler(window time.Duration) func(ctx echo.Context, id string, err error) error {
	retryAfter := int(window.Seconds())

	return func(ctx echo.Context, id string, err error) error {
		ctx.Response().Header().Set("Retry-After", strconv.Itoa(retryAfter))

		sensivity := GetSensivity(ctx)

		switch sensivity{
		case InsignificantEndpoint:
			log.Trace("Request blocked by rate limmiter", request.GetMetadata(ctx))
		case DefaultEndpoint:
			log.Info("Request blocked by rate limmiter", request.GetMetadata(ctx))
		case SensitiveEndpoint:
			log.Warning("Request blocked by rate limmiter", request.GetMetadata(ctx))
		default:
			log.Panic(
				"Invalid use of rateLimiterDenyHandler()",
				"Unknown endpoint sensitivity level",
				request.GetMetadata(ctx),
			)
		}

		return ctx.JSON(
			http.StatusTooManyRequests,
			ResponseBody.Message{
				Message: "Too many requests",
			},
		)
	}
}

func NewRateLimiter() *rateLimiter {
	return new(rateLimiter)
}

func (l *rateLimiter) Max5reqPerHour() echo.MiddlewareFunc {
	window := time.Hour

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate: rate.Every(window / 5),
			Burst: 1,
			ExpiresIn: window * 2,
		}),
		DenyHandler: rateLimiterDenyHandler(window / 5),
		IdentifierExtractor: rateLimiterIdentifierExtractor,
	})
}

func (l *rateLimiter) Max5reqPerMinute() echo.MiddlewareFunc {
	window := time.Minute

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate: rate.Every(window / 5),
			Burst: 3,
			ExpiresIn: window * 2,
		}),
		DenyHandler: rateLimiterDenyHandler(window / 5),
		IdentifierExtractor: rateLimiterIdentifierExtractor,
	})
}

func (l *rateLimiter) Max3reqPerMinute() echo.MiddlewareFunc {
	window := time.Minute

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate: rate.Every(window / 3),
			Burst: 1,
			ExpiresIn: window * 2,
		}),
		DenyHandler: rateLimiterDenyHandler(window / 3),
		IdentifierExtractor: rateLimiterIdentifierExtractor,
	})
}

func (l *rateLimiter) Max1reqPer5Minutes() echo.MiddlewareFunc {
	window := time.Minute

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate: rate.Every(window * 5),
			Burst: 1,
			ExpiresIn: window * 2,
		}),
		DenyHandler: rateLimiterDenyHandler(window * 5),
		IdentifierExtractor: rateLimiterIdentifierExtractor,
	})
}

func (l *rateLimiter) Max1reqPerSecond() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate: 1,
			Burst: 5,
			ExpiresIn: time.Minute,
		}),
		DenyHandler: rateLimiterDenyHandler(1),
		IdentifierExtractor: rateLimiterIdentifierExtractor,
	})
}

func (l *rateLimiter) Max10reqPerSecond() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate: 10,
			Burst: 5,
			ExpiresIn: time.Minute,
		}),
		DenyHandler: rateLimiterDenyHandler(10),
		IdentifierExtractor: rateLimiterIdentifierExtractor,
	})
}

