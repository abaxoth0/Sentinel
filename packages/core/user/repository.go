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
	FindAnyUserByID(string) (*UserDTO.Basic, *Error.Status)

	FindUserByID(string) (*UserDTO.Basic, *Error.Status)

	FindSoftDeletedUserByID(string) (*UserDTO.Basic, *Error.Status)

	FindUserByLogin(string) (*UserDTO.Basic, *Error.Status)

	FindAnyUserByLogin(string) (*UserDTO.Basic, *Error.Status)

	IsLoginExists(login string) (bool, *Error.Status)

    GetRoles(filter *UserDTO.Filter) ([]string, *Error.Status)
}

// Responsible for CUD in CRUD
type repository interface {
	Create(login string, password string) *Error.Status

	SoftDelete(filter *UserDTO.Filter) *Error.Status

	Restore(filter *UserDTO.Filter) *Error.Status

	Drop(filter *UserDTO.Filter) *Error.Status

	DropAllSoftDeleted(filter *UserDTO.Filter) *Error.Status

	ChangeLogin(filter *UserDTO.Filter, newLogin string) *Error.Status

	ChangePassword(filter *UserDTO.Filter, newPassword string) *Error.Status

	ChangeRoles(filter *UserDTO.Filter, newRoles []string) *Error.Status
}

