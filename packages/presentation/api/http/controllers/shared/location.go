package sharedcontroller

import (
	"net"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	LocationProvider "sentinel/packages/common/location"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	controller "sentinel/packages/presentation/api/http/controllers"
)

// If location for session with specified ID already exists - updates this location.
// If there are no location for this session - creates new location for it.
func updateOrCreateLocation(act *ActionDTO.UserTargeted, sessionID string, ip net.IP) *Error.Status {
	controller.Log.Trace("Updating location for session "+sessionID+"...", nil)

	location, err := DB.Database.GetLocationBySessionID(act, sessionID)
	if err != nil {
		if err != Error.StatusNotFound {
			return err
		}

		newLocation, err := LocationProvider.GetLocationFromIP(ip.To4().String())
		if err != nil {
			return err
		}

		newLocation.SessionID = sessionID

		if err := DB.Database.SaveLocation(newLocation); err != nil {
			return err
		}
	} else {
		if config.Debug.Enabled && config.Debug.LocationIP != "" {
			controller.Log.Debug("Request IP changed: "+ip.To4().String()+" -> "+config.Debug.LocationIP, nil)
			ip = net.ParseIP(config.Debug.LocationIP)
		}
		if ip.Equal(location.IP) {
			controller.Log.Info("Location update skipped: IP address of location hasn't changed", nil)
			return nil
		}

		newLocation, err := LocationProvider.GetLocationFromIP(ip.To4().String())
		if err != nil {
			return err
		}

		newLocation.SessionID = sessionID

		if err := DB.Database.UpdateLocation(location.ID, newLocation); err != nil {
			return err
		}
	}

	controller.Log.Trace("Updating location for session "+sessionID+":OK", nil)

	return nil
}

