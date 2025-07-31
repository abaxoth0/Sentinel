package locationtable

import (
	Error "sentinel/packages/common/errors"
	LocationDTO "sentinel/packages/core/location/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/cache"
)

// TODO add private method for updating without authz and add authz for this method, do the same for sessions
// 		(also need to request ActionDTO.Targeted instead of location id)
func (m *Manager) UpdateLocation(id string, newLocation *LocationDTO.Full) *Error.Status {
	log.DB.Trace("Updating locaiton "+id+"...", nil)

	location, err := m.getLocationByID(id)
	if err != nil {
		return err
	}
	if !location.DeletedAt.IsZero() {
		log.DB.Error("Failed to update locaiton "+id, Error.StatusNotFound.Error(), nil)
		return Error.StatusNotFound
	}

	updateQuery := query.New(
		`UPDATE "location" SET
		ip = $1, session_id = $2, country = $3, region = $4, city = $5, latitude = $6, longitude = $7, isp = $8
		WHERE id = $9 AND deleted_at IS NULL;`,
		newLocation.IP,
		newLocation.SessionID,
		newLocation.Country,
		newLocation.Region,
		newLocation.City,
		newLocation.Latitude,
		newLocation.Longitude,
		newLocation.ISP,
		id,
	)

	audit := newAuditDTO(audit.UpdatedOperation, location)

	if err := execTxWithAudit(&audit, updateQuery); err != nil {
		return err
	}

	// TODO handle error
	cache.Client.Delete(
		cache.KeyBase[cache.LocationByID] + id,
		cache.KeyBase[cache.LocationBySessionID] + location.SessionID,
	)

	log.DB.Trace("Updating locaiton "+id+"...", nil)

	return nil
}

