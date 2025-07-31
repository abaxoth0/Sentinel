package controller

import (
	"net"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	LocationProvider "sentinel/packages/common/location"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
)

func UpdateLocation(act *ActionDTO.UserTargeted, sessionID string, ip net.IP) *Error.Status {
	Log.Trace("Updating location for session "+sessionID+"...", nil)

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
			Log.Debug("Request IP changed: "+ip.To4().String()+" -> "+config.Debug.LocationIP, nil)
			ip = net.ParseIP(config.Debug.LocationIP)
		}
		if ip.Equal(location.IP) {
			Log.Info("Location update skipped: IP address of location hasn't changed", nil)
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

	Log.Trace("Updating location for session "+sessionID+":OK", nil)

	return nil
}

