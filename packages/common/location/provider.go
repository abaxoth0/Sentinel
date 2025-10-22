package locationprovider

import (
	"context"
	"net"
	"net/http"
	"sentinel/packages/common/config"
	"sentinel/packages/common/encoding/json"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/structs"
	LocationDTO "sentinel/packages/core/location/DTO"
	"time"

	"github.com/sony/gobreaker/v2"
)

var log = logger.NewSource("LOCATION PROVIDER", logger.Default)

const fields string = "?fields=status,message,country,countryCode,region,regionName,city,lat,lon,isp"

type geoIpResponseBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	LocationDTO.Full `json:",inline"`
}

var circuitBreaker = gobreaker.NewCircuitBreaker[*LocationDTO.Full](gobreaker.Settings{
	Name:        "Location provider",
	Interval:    time.Second,
	Timeout:     time.Second * 20,
	MaxRequests: 10,
})

var requestTimeout = time.Second * 5

// Returns user location in raw string format based on specified ip address
func GetLocationFromIP(ip string) (*LocationDTO.Full, *Error.Status) {
	log.Trace("Getting location for "+ip+"...", nil)

	if config.Debug.Enabled && config.Debug.LocationIP != "" && ip != config.Debug.LocationIP {
		log.Debug("IP changed: "+ip+" -> "+config.Debug.LocationIP, nil)
		ip = config.Debug.LocationIP
	}

	dto, err := circuitBreaker.Execute(func() (*LocationDTO.Full, error) {
		var res *http.Response
		var err error

		structs.SetTimeout(context.Background(), requestTimeout, func(ctx context.Context) {
			res, err = http.Get("http://ip-api.com/json/" + ip + fields)
		})
		if err != nil {
			log.Error("Failed to get location for "+ip, err.Error(), nil)
			return nil, Error.NewStatusError(
				"Failed to get user location:"+err.Error(),
				http.StatusInternalServerError,
			)
		}
		defer res.Body.Close()

		body, err := json.Decode[geoIpResponseBody](res.Body)
		if err != nil {
			return nil, Error.NewStatusError(
				"Failed to read response body from location provider",
				http.StatusInternalServerError,
			)
		}
		if body.Status != "success" {
			log.Error("Failed to get location for "+ip, body.Message, nil)
			return nil, Error.NewStatusError(
				"Failed to read response body from location provider",
				http.StatusInternalServerError,
			)
		}

		body.Full.CreatedAt = time.Now()
		body.Full.IP = net.ParseIP(ip)

		return &body.Full, nil
	})

	if err != nil {
		if e, ok := err.(*Error.Status); ok {
			return nil, e
		}
		log.Error(
			"Request to location provider has been blocked by circuit breaker",
			err.Error(),
			nil,
		)
		return nil, Error.StatusInternalError
	}

	log.Trace("Getting location for "+ip+": OK", nil)

	return dto, nil
}
