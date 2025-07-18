package locationtable

import (
	Error "sentinel/packages/common/errors"
	LocationDTO "sentinel/packages/core/location/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
)

// TODO add private method for updating without authz and add authz for this method, do the same for sessions
// 		(also need to request ActionDTO.Targeted instead of location id)
func (_ *Manager) UpdateLocation(id string, newLocation *LocationDTO.Full) *Error.Status {
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
	return executor.Exec(connection.Primary, updateQuery)
}

