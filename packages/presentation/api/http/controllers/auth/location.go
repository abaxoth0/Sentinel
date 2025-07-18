package authcontroller

import (
	Error "sentinel/packages/common/errors"
	LocationProvider "sentinel/packages/common/location"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	controller "sentinel/packages/presentation/api/http/controllers"
)

func updateLocation(act *ActionDTO.UserTargeted, sessionID string, ip string) *Error.Status {
	controller.Logger.Trace("Updating location for session "+sessionID+"...", nil)

	location, err := DB.Database.GetLocationBySessionID(act, sessionID)
	if err != nil {
		if err != Error.StatusNotFound {
			return err
		}

		newLocation, err := LocationProvider.GetLocationFromIP(ip)
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
		newLocation, err := LocationProvider.GetLocationFromIP(ip)
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

