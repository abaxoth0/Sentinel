package user

import (
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/common/errors"
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

	IsLoginAvailable(login string) (bool, *Error.Status)

    GetRoles(filter *UserDTO.Filter) ([]string, *Error.Status)
}

// Responsible for CUD in CRUD
type repository interface {
    // Returns user id if error is nil, otherwise returns empty string and error
	Create(login string, password string) (string, *Error.Status)

	SoftDelete(filter *UserDTO.Filter) *Error.Status

	Restore(filter *UserDTO.Filter) *Error.Status

	Drop(filter *UserDTO.Filter) *Error.Status

	DropAllSoftDeleted(filter *UserDTO.Filter) *Error.Status

	ChangeLogin(filter *UserDTO.Filter, newLogin string) *Error.Status

	ChangePassword(filter *UserDTO.Filter, newPassword string) *Error.Status

	ChangeRoles(filter *UserDTO.Filter, newRoles []string) *Error.Status

    Activate(token string) *Error.Status
}

