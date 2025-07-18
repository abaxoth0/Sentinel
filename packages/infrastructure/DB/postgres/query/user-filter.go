package query

import (
	"fmt"
	"sentinel/packages/core/filter"
	"sentinel/packages/core/user"
	"strconv"
)
type UserFilter struct {
	Property  user.Property
	Cond 	  filterCond
	Value 	  any
}

func (f UserFilter) Build(n int) string {
	base := string(f.Property) + " " + string(f.Cond)

	// SELECT ... FROM ... WHERE <property> <cond> $n (omit if cond is "IS NULL" or "IS NOT NULL")
	if f.Cond == CondIsNull || f.Cond == CondIsNotNull {
		return base
	}

	return base + " $" + strconv.FormatInt(int64(n), 10)
}

func (f UserFilter) StringValue() string {
	v, ok := f.Value.(string)
	if !ok {
		queryLogger.Panic(
			"Failed to find user",
			fmt.Sprintf("Query filter field 'value' has invalid type. Expected string, but got %T", f.Value),
			nil,
		)
		return ""
	}

	return v
}

// Converts []filter.Entity[user.Property] to []QueryFilter.
// Returns error if there are some validation error.
func MapAndValidateUserFilters(filters []filter.Entity[user.Property]) ([]UserFilter, error) {
	queryFilter := make([]UserFilter, len(filters))

	for i, f := range filters {
		if f.Property == "" || f.Cond == 0 {
			return nil, fmt.Errorf(
				"Filter property or condition has invalid value. Property value: %v; Condition value: %v",
				f.Property, f.Cond,
			)
		}
		if f.Cond != filter.IsNull && f.Cond != filter.IsNotNull && f.Value == nil{
			return nil, fmt.Errorf("Filter value is missing or nil")
		}

		queryFilter[i] = UserFilter{
			Property: f.Property,
			Cond: condMap[f.Cond],
			Value: f.Value,
		}
	}

	return queryFilter, nil
}

