package locationtable

import (
	Error "sentinel/packages/common/errors"
	LocationDTO "sentinel/packages/core/location/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"

	"github.com/google/uuid"
)

func (_ *Manager) SaveLocation(dto *LocationDTO.Full) *Error.Status {
	log.DB.Trace("Saving location...", nil)

	insertQuery := query.New(
		`INSERT INTO "location" (id, ip, session_id, country, region, city, latitude, longitude, isp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`,
		uuid.NewString(),
		dto.IP,
		dto.SessionID,
		dto.Country,
		dto.Region,
		dto.City,
		dto.Latitude,
		dto.Longitude,
		dto.ISP,
		dto.CreatedAt,
	)

	if err := executor.Exec(connection.Primary, insertQuery); err != nil {
		return err
	}

	log.DB.Trace("Saving location: OK", nil)

	return nil
}
