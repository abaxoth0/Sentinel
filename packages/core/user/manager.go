package user

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
)

type Manager interface {
	creator
	seeker
	updater
	deleter
}

type creator interface {
	// Returns user id if error is nil, otherwise returns empty string and error
	Create(login string, password string) (string, *Error.Status)
}

type seeker interface {
	SearchUsers(act *ActionDTO.Basic, rawFilters []string, page int, pageSize int) ([]*UserDTO.Public, *Error.Status)

	GetUserByID(string) (*UserDTO.Full, *Error.Status)

	GetSoftDeletedUserByID(string) (*UserDTO.Full, *Error.Status)

	GetUserByLogin(string) (*UserDTO.Full, *Error.Status)

	GetUserBySessionID(string) (*UserDTO.Full, *Error.Status)

	IsLoginInUse(login string) bool

	GetRoles(act *ActionDTO.UserTargeted) ([]string, *Error.Status)

	GetUserVersion(UID string) (uint32, *Error.Status)
}

type updater interface {
	ChangeLogin(act *ActionDTO.UserTargeted, newLogin string) *Error.Status

	ChangePassword(act *ActionDTO.UserTargeted, newPassword string) *Error.Status

	ChangeRoles(act *ActionDTO.UserTargeted, newRoles []string) *Error.Status

	Activate(token string) *Error.Status
}

type deleter interface {
	SoftDelete(act *ActionDTO.UserTargeted) *Error.Status

	Restore(act *ActionDTO.UserTargeted, newLogin string) *Error.Status

	Drop(act *ActionDTO.UserTargeted) *Error.Status

	DropAllSoftDeleted(act *ActionDTO.Basic) *Error.Status

	BulkSoftDelete(act *ActionDTO.Basic, UIDs []string) *Error.Status

	BulkRestore(act *ActionDTO.Basic, UIDs []string) *Error.Status
}
