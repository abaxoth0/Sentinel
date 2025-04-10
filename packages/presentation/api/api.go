package api

import (
	"sentinel/packages/common/config"
	"sentinel/packages/common/util"
)

// Returns auto generated base URL of this service (based on config)
//
// Exmaple: http://localhost:1234
func GetBaseURL() string {
    transport := util.Ternary(config.HTTP.Secured, "https", "http")

    return transport + "://" + config.HTTP.Domain + ":" + config.HTTP.Port
}

