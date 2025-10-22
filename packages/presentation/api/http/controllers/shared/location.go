package sharedcontroller

import (
	"net"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	LocationProvider "sentinel/packages/common/location"
	ActionDTO "sentinel/packages/core/action/DTO"
	LocationDTO "sentinel/packages/core/location/DTO"
	"sentinel/packages/infrastructure/DB"
	controller "sentinel/packages/presentation/api/http/controllers"
)

// If location for session with specified ID already exists - updates this location.
// If there are no location for this session - creates new location for it.
func updateOrCreateLocation(act *ActionDTO.UserTargeted, sessionID string, ip net.IP) (*LocationDTO.Full, *Error.Status) {
	controller.Log.Trace("Updating location for session "+sessionID+"...", nil)

	var err *Error.Status
	var newLocation *LocationDTO.Full

	location, err := DB.Database.GetLocationBySessionID(act, sessionID)
	if err != nil {
		if err != Error.StatusNotFound {
			return nil, err
		}

		newLocation, err = LocationProvider.GetLocationFromIP(ip.To4().String())
		if err != nil {
			return nil, err
		}

		newLocation.SessionID = sessionID

		if err := DB.Database.SaveLocation(newLocation); err != nil {
			return nil, err
		}
	} else {
		if config.Debug.Enabled && config.Debug.LocationIP != "" {
			controller.Log.Debug("Request IP changed: "+ip.To4().String()+" -> "+config.Debug.LocationIP, nil)
			ip = net.ParseIP(config.Debug.LocationIP)
		}
		if ip.Equal(location.IP) {
			controller.Log.Info("Location update skipped: IP address of location hasn't changed", nil)
			return nil, nil
		}

		newLocation, err = LocationProvider.GetLocationFromIP(ip.To4().String())
		if err != nil {
			return nil, err
		}

		newLocation.SessionID = sessionID

		if err := DB.Database.UpdateLocation(location.ID, newLocation); err != nil {
			return nil, err
		}
	}

	controller.Log.Trace("Updating location for session "+sessionID+":OK", nil)

	return newLocation, nil
}
