package locationtable

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	LocationDTO "sentinel/packages/core/location/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/dblog"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
)

func (_ *Manager) getLocationByID(id string) (*LocationDTO.Full, *Error.Status) {
	dblog.Logger.Trace("Getting location "+id+"...", nil)

	selectQuery := query.New(
		`SELECT id, ip, session_id, country, region, city, latitude, longitude, isp, deleted_at, created_at
		FROM "location" WHERE id = $1`,
		id,
	)

	dto, err := executor.FullLocationDTO(
		connection.Replica,
		selectQuery,
		cache.KeyBase[cache.LocationByID]+id,
	)
	if err != nil {
		return nil, err
	}

	dblog.Logger.Trace("Getting location "+id+": OK", nil)

	return dto, nil
}

func (l *Manager) GetLocationByID(act *ActionDTO.UserTargeted, id string) (*LocationDTO.Full, *Error.Status) {
	if act.TargetUID != act.RequesterUID {
		if err := authz.User.GetSessionLocation(act.RequesterRoles); err != nil {
			return nil, err
		}
	}

	return l.getLocationByID(id)
}

func (l *Manager) GetLocationBySessionID(act *ActionDTO.UserTargeted, sessionID string) (*LocationDTO.Full, *Error.Status) {
	dblog.Logger.Trace("Getting location for session "+sessionID+"...", nil)

	if act.TargetUID != act.RequesterUID {
		if err := authz.User.GetSessionLocation(act.RequesterRoles); err != nil {
			return nil, err
		}
	}

	selectQuery := query.New(
		`SELECT id, ip, session_id, country, region, city, latitude, longitude, isp, deleted_at, created_at
		FROM "location" WHERE session_id = $1`,
		sessionID,
	)

	dto, err := executor.FullLocationDTO(
		connection.Replica,
		selectQuery,
		cache.KeyBase[cache.LocationBySessionID]+sessionID,
	)
	if err != nil {
		return nil, err
	}

	dblog.Logger.Trace("Getting location for session "+sessionID+": OK", nil)

	return dto, nil
}
