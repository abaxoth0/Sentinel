package locationtable

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
)

func (l *Manager) deleteLocation(id string, act *ActionDTO.Targeted, drop bool) *Error.Status {
	if act.TargetUID != act.RequesterUID {
		if err := authz.User.DeleteLocation(act.RequesterRoles); err != nil {
			return err
		}
	}

	location, err := l.getLocationByID(id)
	if err != nil {
		return err
	}

	var stateUpdateQuery *query.Query

	if drop {
		if !location.DeletedAt.IsZero() {
			return Error.NewStatusError(
				"Only soft deleted locations can be dropped",
				http.StatusBadRequest,
			)
		}
		stateUpdateQuery = query.New(
			`DELETE FROM "location" WHERE id = $1 AND deleted_at IS NOT NULL;`,
			id,
		)
	} else {
		stateUpdateQuery = query.New(
			`UPDATE "location" SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;`,
			id,
		)
	}

	return executor.Exec(connection.Primary, stateUpdateQuery)
}

func (l *Manager) SoftDeleteLocation(id string, act *ActionDTO.Targeted) *Error.Status {
	return l.deleteLocation(id, act, false)
}

func (l *Manager) DropLocation(id string, act *ActionDTO.Targeted) *Error.Status {
	return l.deleteLocation(id, act, true)
}

