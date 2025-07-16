package postgres

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	LocationDTO "sentinel/packages/core/location/DTO"
	"sentinel/packages/infrastructure/auth/authz"

	"github.com/google/uuid"
)

type location struct {
	//
}

// TODO add cache
// TODO add audit

func (_ *location) SaveLocation(dto *LocationDTO.Full) *Error.Status {
	query := newQuery(
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

	return query.Exec(primaryConnection)
}

func (_ *location) getLocationByID(id string) (*LocationDTO.Full, *Error.Status) {
	query := newQuery(
		`SELECT id, ip, session_id, country, region, city, latitude, longitude, isp, deleted_at, created_at
		FROM "location" WHERE id = $1`,
		id,
	)

	dto, err := query.FullLocationDTO(replicaConnection)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (l *location) GetLocationByID(act *ActionDTO.Targeted, id string) (*LocationDTO.Full, *Error.Status) {
	// TODO add self targeted action? (e.g. banned_user won't be able to see his sessions)
	//		(for all this kind of actions)
	if act.TargetUID != act.RequesterUID {
		if err := authz.User.GetSessionLocation(act.RequesterRoles); err != nil {
			return nil, err
		}
	}

	return l.getLocationByID(id)
}

func (l *location) GetLocationBySessionID(act *ActionDTO.Targeted, sessionID string) (*LocationDTO.Full, *Error.Status) {
	// TODO add self targeted action? (e.g. banned_user won't be able to see his sessions)
	//		(for all this kind of actions)
	if act.TargetUID != act.RequesterUID {
		if err := authz.User.GetSessionLocation(act.RequesterRoles); err != nil {
			return nil, err
		}
	}

	query := newQuery(
		`SELECT id, ip, session_id, country, region, city, latitude, longitude, isp, deleted_at, created_at
		FROM "location" WHERE session_id = $1`,
		sessionID,
	)

	dto, err := query.FullLocationDTO(replicaConnection)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

// TODO add private method for updating without authz and add authz for this method, do the same for sessions
// 		(also need to request ActionDTO.Targeted instead of location id)
func (_ *location) UpdateLocation(id string, newLocation *LocationDTO.Full) *Error.Status {
	query := newQuery(
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
	return query.Exec(primaryConnection)
}

func (l *location) deleteLocation(id string, act *ActionDTO.Targeted, drop bool) *Error.Status {
	if act.TargetUID != act.RequesterUID {
		if err := authz.User.DeleteLocation(act.RequesterRoles); err != nil {
			return err
		}
	}

	location, err := l.getLocationByID(id)
	if err != nil {
		return err
	}

	var query *query

	if drop {
		if !location.DeletedAt.IsZero() {
			return Error.NewStatusError(
				"Only soft deleted locations can be dropped",
				http.StatusBadRequest,
			)
		}
		query = newQuery(
			`DELETE FROM "location" WHERE id = $1 AND deleted_at IS NOT NULL;`,
			id,
		)
	} else {
		query = newQuery(
			`UPDATE "location" SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;`,
			id,
		)
	}

	return query.Exec(primaryConnection)
}

func (l *location) SoftDeleteLocation(id string, act *ActionDTO.Targeted) *Error.Status {
	return l.deleteLocation(id, act, false)
}

func (l *location) DropLocation(id string, act *ActionDTO.Targeted) *Error.Status {
	return l.deleteLocation(id, act, true)
}

