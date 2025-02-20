package postgres

import (
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errs"
)

type repository struct {
    //
}

func (_ *repository) Create(login string, password string) (string, error) {
    return "", nil
}

func (_ *repository) SoftDelete(filter *UserDTO.Filter) *Error.Status {
    return nil
}

func (_ *repository) Restore(filter *UserDTO.Filter) *Error.Status {
    return nil
}

func (_ *repository) Drop(filter *UserDTO.Filter) *Error.Status {
    return nil
}

func (_ *repository) DropAllSoftDeleted(requesterRoles []string) *Error.Status {
    return nil
}

func (_ *repository) ChangeLogin(filter *UserDTO.Filter, newLogin string) *Error.Status {
    return nil
}

func (_ *repository) ChangePassword(filter *UserDTO.Filter, newPassword string) *Error.Status {
    return nil
}

func (_ *repository) ChangeRoles(filter *UserDTO.Filter, newRoles []string) *Error.Status {
    return nil
}

func (_ *repository) GetRoles(filter *UserDTO.Filter) ([]string, *Error.Status) {
    return []string{}, nil
}

