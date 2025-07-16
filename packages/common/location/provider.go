package locationprovider

import (
	"net"
	"net/http"
	"sentinel/packages/common/config"
	"sentinel/packages/common/encoding/json"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	LocationDTO "sentinel/packages/core/location/DTO"
	"time"
)

var locationLogger = logger.NewSource("LOCATION", logger.Default)

const fields string = "?fields=status,message,country,countryCode,region,regionName,city,lat,lon,isp"

type geoIpResponseBody struct {
	Status 	string 		`json:"status"`
	Message string 		`json:"message"`

	LocationDTO.Full 	`json:",inline"`
}

// Returns user location in raw string format based on specified ip address
func GetLocationFromIP(ip string) (*LocationDTO.Full, *Error.Status) {
	if config.Debug.Enabled && config.Debug.LocationIP != "" {
		locationLogger.Debug("IP changed from "+ip+" to "+config.Debug.LocationIP, nil)
		ip = config.Debug.LocationIP
	}

	locationLogger.Trace("Getting location for "+ip+"...", nil)

	res, err := http.Get("http://ip-api.com/json/"+ip+fields)
	if err != nil {
		locationLogger.Error("Failed to get location for "+ip, err.Error(), nil)
		return nil, Error.NewStatusError(
			"Failed to get user location:" + err.Error(),
			http.StatusInternalServerError,
		)
	}
	defer res.Body.Close()

	body, err := json.Decode[geoIpResponseBody](res.Body)
	if err != nil {
		locationLogger.Error("Failed to get location for "+ip, err.Error(), nil)
		return nil, Error.NewStatusError(
			"Failed to read response body from location provider",
			http.StatusInternalServerError,
		)
	}
	if body.Status != "success" {
		locationLogger.Error("Failed to get location for "+ip, body.Message, nil)
		return nil, Error.NewStatusError(
			"Failed to read response body from location provider",
			http.StatusInternalServerError,
		)
	}

	body.Full.CreatedAt = time.Now()
	body.Full.IP = net.ParseIP(ip)

	locationLogger.Trace("Getting location for "+ip+": OK", nil)

	return &body.Full, nil
}

