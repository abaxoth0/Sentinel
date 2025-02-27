package user

import (
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
)

type Repository interface {
	seeker
	repository
}

// Responsible for R in CRUD
type seeker interface {
	FindUserByID(string) (*UserDTO.Indexed, *Error.Status)

	FindSoftDeletedUserByID(string) (*UserDTO.Indexed, *Error.Status)

	FindUserByLogin(string) (*UserDTO.Indexed, *Error.Status)

	IsLoginExists(login string) (bool, *Error.Status)
}

// Responsible for CUD in CRUD
type repository interface {
	// Returns uid and error
	Create(login string, password string) (*Error.Status)

	SoftDelete(filter *UserDTO.Filter) *Error.Status

	Restore(filter *UserDTO.Filter) *Error.Status

	Drop(filter *UserDTO.Filter) *Error.Status

	DropAllSoftDeleted(requesterRoles []string) *Error.Status

	ChangeLogin(filter *UserDTO.Filter, newLogin string) *Error.Status

	ChangePassword(filter *UserDTO.Filter, newPassword string) *Error.Status

	ChangeRoles(filter *UserDTO.Filter, newRoles []string) *Error.Status

	GetRoles(filter *UserDTO.Filter) ([]string, *Error.Status)
}

