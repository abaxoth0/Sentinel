package userfilterparser

import (
	"errors"
	"sentinel/packages/core/filter"
	"sentinel/packages/core/user"
	FilterMapper "sentinel/packages/infrastructure/mappers/filter"
)

var validPropertyCondMap = map[user.Property]map[filter.Condition]bool{
	user.IdProperty: {
		filter.Equal: true,
	},
	user.LoginProperty: {
		filter.Equal: true,
		filter.Like:  true,
	},
	user.PasswordProperty: {
		filter.Equal: true,
	},
	user.RolesProperty: {
		filter.Contains: true,
		filter.Containd: true,
	},
	user.DeletedAtProperty: {
		filter.IsNull:         true,
		filter.IsNotNull:      true,
		filter.Equal:          true,
		filter.Less:           true,
		filter.LessOrEqual:    true,
		filter.Greater:        true,
		filter.GreaterOrEqual: true,
	},
}

// Validates if 'cond' can be applied to 'property'.
// If you want to know valid conds for each property see 'validPropertyCondMap'.
func validatePropertyCond(property user.Property, cond filter.Condition) error {
	validConds := validPropertyCondMap[property]
	if !validConds[cond] {
		// cond is valid, so there are no cases when error will be not nil
		condStr, _ := FilterMapper.FormatCond(cond)

		return errors.New("Invalid condition '" + condStr + "' for user property '" + string(property) + "'")
	}
	return nil
}
