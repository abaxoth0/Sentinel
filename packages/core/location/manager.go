package location

import (
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	LocationDTO "sentinel/packages/core/location/DTO"
)

type Manager interface {
	creator
	seeker
	updater
	deleter
}

type creator interface {
	SaveLocation(dto *LocationDTO.Full) *Error.Status
}

type seeker interface {
	GetLocationByID(act *ActionDTO.Targeted, id string) (*LocationDTO.Full, *Error.Status)
	GetLocationBySessionID(act *ActionDTO.Targeted, id string) (*LocationDTO.Full, *Error.Status)
}

type updater interface {
	UpdateLocation(id string, newLocation *LocationDTO.Full) *Error.Status
}

type deleter interface {
	SoftDeleteLocation(id string, act *ActionDTO.Targeted) *Error.Status
	DropLocation(id string, act *ActionDTO.Targeted) *Error.Status
}

