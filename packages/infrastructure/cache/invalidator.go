package cache

import (
	Error "sentinel/packages/common/errors"
	UserDTO "sentinel/packages/core/user/DTO"
	"slices"
)

// Used to automically find/delete invalid cache keys
type Invalidator interface {
	GetInvalidKeys() []string
	Invalidate() *Error.Status
}

// Used to automically find/delete invalid cache keys by analyzing
// changes between 'old' and 'current' version of specified DTO.
type BasicUserDtoInvalidator struct {
	old 		 	*UserDTO.Full
	current 	 	*UserDTO.Full
	invalidKeys	 	[]string
	isIdValid 	 	bool
	isLoginValid 	bool
	isOldDeleted 	bool
	isVersionValid 	bool
}

func NewBasicUserDtoInvalidator(old *UserDTO.Full, current *UserDTO.Full) *BasicUserDtoInvalidator {
	return &BasicUserDtoInvalidator{
		old: old,
		current: current,
		invalidKeys: []string{},
		isIdValid: true,
		isLoginValid: true,
		isOldDeleted: !old.DeletedAt.IsZero(),
		isVersionValid: true,
	}
}

// Invalidates id cache keys of 'old'.
// Early returns if they was already invalidated.
func (i *BasicUserDtoInvalidator) invalidateIdKeys() {
	if !i.isIdValid {
		return
	}

	i.invalidKeys = append(i.invalidKeys, KeyBase[AnyUserById] + i.old.ID)
	i.invalidKeys = append(i.invalidKeys, KeyBase[UserVersionByID] + i.old.ID)

	if !i.old.DeletedAt.Equal(*i.current.DeletedAt) {
		i.invalidKeys = append(i.invalidKeys, KeyBase[UserById] + i.old.ID)
		i.invalidKeys = append(i.invalidKeys, KeyBase[DeletedUserById] + i.old.ID)
		i.isIdValid = false
		return
	}

	if !i.isOldDeleted {
		i.invalidKeys = append(i.invalidKeys, KeyBase[UserById] + i.old.ID)
	} else {
		i.invalidKeys = append(i.invalidKeys, KeyBase[DeletedUserById] + i.old.ID)
	}

	i.isIdValid = false
}

// Invalidates login cache keys of 'old'.
// Early returns if they was already invalidated.
func (i *BasicUserDtoInvalidator) invalidateLoginKeys() {
	if !i.isLoginValid {
		return
	}

	i.invalidKeys = append(i.invalidKeys, KeyBase[AnyUserByLogin] + i.old.Login)

	if !i.isOldDeleted || !i.old.DeletedAt.Equal(*i.current.DeletedAt) {
		i.invalidKeys = append(i.invalidKeys, KeyBase[UserByLogin] + i.old.Login)
	}

	i.isLoginValid = false
}

func (i *BasicUserDtoInvalidator) GetInvalidKeys() []string {
	isDeletedAtChaged := !i.old.DeletedAt.Equal(*i.current.DeletedAt)

	if i.old.ID != i.current.ID ||
	i.old.Login != i.current.Login ||
	i.old.Password != i.current.Password ||
	i.old.Version != i.current.Version ||
	isDeletedAtChaged {
		i.invalidateIdKeys()
		i.invalidateLoginKeys()
	}
	if !slices.Equal(i.old.Roles, i.current.Roles) || isDeletedAtChaged {
		i.invalidateIdKeys()
		i.invalidateLoginKeys()
		i.invalidKeys = append(i.invalidKeys, KeyBase[UserRolesById] + i.old.ID)
	}
	return i.invalidKeys
}

// Deletes all invalid keys of 'old' from cache
func (i *BasicUserDtoInvalidator) Invalidate() *Error.Status {
	return Client.Delete(i.GetInvalidKeys()...)
}

