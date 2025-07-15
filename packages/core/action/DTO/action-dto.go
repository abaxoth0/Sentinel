package actiondto

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
)

type Any interface {
	ValidateRequesterUID() *Error.Status
}

type Basic struct {
	RequesterUID   string
	RequesterRoles []string
}

func NewBasic(requesterUID string, requestedRoles []string) *Basic {
    return &Basic{
        RequesterUID: requesterUID,
        RequesterRoles: requestedRoles,
    }
}

func (dto *Basic) ValidateRequesterUID() *Error.Status {
    if err := validation.UUID(dto.RequesterUID); err != nil {
        return err.ToStatus(
            "Requester user ID is not specified",
            "Invalid requester user ID",
            )
    }
    return nil
}

func (dto *Basic) ToTargeted(targetUID string) *Targeted {
	return &Targeted{
		TargetUID: targetUID,
		Basic: *dto,
	}
}

type Targeted struct {
	TargetUID string
    Basic
}

func NewTargeted(targetdUID string, requesterUID string, requestedRoles []string) *Targeted {
    return &Targeted{
        TargetUID: targetdUID,
        Basic: Basic{
            RequesterUID: requesterUID,
            RequesterRoles: requestedRoles,
        },
    }
}

func (dto *Targeted) ValidateTargetUID() *Error.Status {
    if err := validation.UUID(dto.TargetUID); err != nil {
        return err.ToStatus(
            "Target user ID is not specified",
            "Invalid target user ID",
        )
    }
    return nil
}

func (dto *Targeted) ValidateUIDs() *Error.Status {
    if err := dto.ValidateTargetUID(); err != nil {
        return err
    }
    if err := dto.ValidateRequesterUID(); err != nil {
        return err
    }
    return nil
}

