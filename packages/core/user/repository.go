package user

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
)

type Repository interface {
	creator
	seeker
	updater
	deleter
}

type creator interface {
	// Returns user id if error is nil, otherwise returns empty string and error
	Create(login string, password string) (string, *Error.Status)

	SearchUsers(act *ActionDTO.Basic, rawFilters []string) ([]*UserDTO.Public, *Error.Status)
}

type seeker interface {
	FindAnyUserByID(string) (*UserDTO.Basic, *Error.Status)

	FindUserByID(string) (*UserDTO.Basic, *Error.Status)

	FindSoftDeletedUserByID(string) (*UserDTO.Basic, *Error.Status)

	FindUserByLogin(string) (*UserDTO.Basic, *Error.Status)

	FindAnyUserByLogin(string) (*UserDTO.Basic, *Error.Status)

	IsLoginAvailable(login string) bool

    GetRoles(act *ActionDTO.Targeted) ([]string, *Error.Status)
}

type updater interface {
	ChangeLogin(act *ActionDTO.Targeted, newLogin string) *Error.Status

	ChangePassword(act *ActionDTO.Targeted, newPassword string) *Error.Status

	ChangeRoles(act *ActionDTO.Targeted, newRoles []string) *Error.Status

	Activate(token string) *Error.Status
}

type deleter interface {
	SoftDelete(act *ActionDTO.Targeted) *Error.Status

	Restore(act *ActionDTO.Targeted) *Error.Status

	Drop(act *ActionDTO.Targeted) *Error.Status

	DropAllSoftDeleted(act *ActionDTO.Basic) *Error.Status
}

