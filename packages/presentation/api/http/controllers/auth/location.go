package authcontroller

import (
	"net"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	LocationProvider "sentinel/packages/common/location"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	controller "sentinel/packages/presentation/api/http/controllers"
)

func updateLocation(act *ActionDTO.UserTargeted, sessionID string, ip net.IP) *Error.Status {
	controller.Logger.Trace("Updating location for session "+sessionID+"...", nil)

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

		controller.Logger.Trace("Saving new location for session "+sessionID+"...", nil)

		if err := DB.Database.SaveLocation(newLocation); err != nil {
			return err
		}

		controller.Logger.Trace("Saving new location for session "+sessionID+": OK", nil)
	} else {
		if config.Debug.Enabled && config.Debug.LocationIP != "" {
			controller.Logger.Debug("Request IP changed: "+ip.To4().String()+" -> "+config.Debug.LocationIP, nil)
			ip = net.ParseIP(config.Debug.LocationIP)
		}
		if ip.Equal(location.IP) {
			controller.Logger.Trace("Location update skipped: location IP and request IP are the same", nil)
			return nil
		}

		newLocation, err := LocationProvider.GetLocationFromIP(ip.To4().String())
		if err != nil {
			return err
		}

		newLocation.SessionID = sessionID

		controller.Logger.Trace("Updating existing location for session "+sessionID+"...", nil)

		if err := DB.Database.UpdateLocation(location.ID, newLocation); err != nil {
			return err
		}

		controller.Logger.Trace("Updating existing location for session "+sessionID+": OK", nil)
	}

	controller.Logger.Trace("Updating location for session "+sessionID+":OK", nil)

	return nil
}

