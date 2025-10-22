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
	Reason         string
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

func (dto *Basic) ToUserTargeted(targetUID string) *UserTargeted {
	return &UserTargeted{
		TargetUID: targetUID,
		Basic:     *dto,
	}
}

type UserTargeted struct {
	TargetUID string
	Basic
}

func NewUserTargeted(targetdUID string, requesterUID string, requestedRoles []string) *UserTargeted {
	return &UserTargeted{
		TargetUID: targetdUID,
		Basic: Basic{
			RequesterUID:   requesterUID,
			RequesterRoles: requestedRoles,
		},
	}
}

func (dto *UserTargeted) ValidateTargetUID() *Error.Status {
	if err := validation.UUID(dto.TargetUID); err != nil {
		return err.ToStatus(
			"Target user ID is not specified",
			"Invalid target user ID",
		)
	}
	return nil
}

func (dto *UserTargeted) ValidateUIDs() *Error.Status {
	if err := dto.ValidateTargetUID(); err != nil {
		return err
	}
	if err := dto.ValidateRequesterUID(); err != nil {
		return err
	}
	return nil
}
