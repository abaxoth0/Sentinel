package activation

import (
	Error "sentinel/packages/common/errors"
	ActivationDTO "sentinel/packages/core/activation/DTO"
)

type Repository interface {
    seeker
    repository
}

type repository interface {
    Activate(token string) *Error.Status
    Reactivate(login string) *Error.Status
}

type seeker interface {
    FindActivationByToken(token string) (*ActivationDTO.Basic, *Error.Status)
    FindActivationByUserLogin(login string) (*ActivationDTO.Basic, *Error.Status)
}

