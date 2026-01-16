package locationtable

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	"sentinel/packages/infrastructure/DB/postgres/dblog"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
)

func (l *Manager) deleteLocation(id string, act *ActionDTO.UserTargeted, drop bool) *Error.Status {
	logPrefix := util.Ternary(drop, "Hard ", "Soft ")

	dblog.Logger.Info(logPrefix+"deleting location "+id+"...", nil)

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

	audit := newAuditDTO(audit.DeleteOperation, location)

	if err := execTxWithAudit(&audit, stateUpdateQuery); err != nil {
		return err
	}

	cache.Client.Delete(
		cache.KeyBase[cache.LocationByID]+id,
		cache.KeyBase[cache.LocationBySessionID]+location.SessionID,
	)

	dblog.Logger.Info(logPrefix+"deleting location "+id+": OK", nil)

	return nil
}

func (l *Manager) SoftDeleteLocation(id string, act *ActionDTO.UserTargeted) *Error.Status {
	return l.deleteLocation(id, act, false)
}

func (l *Manager) DropLocation(id string, act *ActionDTO.UserTargeted) *Error.Status {
	return l.deleteLocation(id, act, true)
}
