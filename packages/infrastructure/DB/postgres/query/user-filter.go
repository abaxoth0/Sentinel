package query

import (
	"errors"
	"fmt"
	"sentinel/packages/core/filter"
	"sentinel/packages/core/user"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"strconv"
)

type UserFilter struct {
	Property user.Property
	Cond     filterCond
	Value    any
}

// Creates valid SQL query condition based on this UserFilter.
// n - number of argument, used for value placeholders.
// If Cond is CondIsNull or CondIsNotNull then placeholder will be ommited.
// Example output:
//   - "id = $1"
//   - "roles @> $8"
//   - "deleted_at IS NULL"
func (f UserFilter) Build(n int) string {
	base := string(f.Property) + " " + string(f.Cond)

	// SELECT ... FROM ... WHERE <property> <cond> $n (omit if cond is "IS NULL" or "IS NOT NULL")
	if f.Cond == CondIsNull || f.Cond == CondIsNotNull {
		return base
	}

	return base + " $" + strconv.FormatInt(int64(n), 10)
}

// Converts []filter.Entity[user.Property] to []QueryFilter.
func MapUserFilters(filters []filter.Entity[user.Property]) ([]UserFilter, error) {
	log.DB.Trace("Mapping user filters...", nil)

	queryFilter := make([]UserFilter, len(filters))

	for i, f := range filters {
		if f.Property == "" || f.Cond == 0 {
			errMsg := fmt.Sprintf(
				"Filter property or condition has invalid value. Property: %v; Condition: %v",
				f.Property, f.Cond,
			)
			log.DB.Error("Failed to map user filters", errMsg, nil)
			return nil, errors.New(errMsg)
		}
		if f.Cond != filter.IsNull && f.Cond != filter.IsNotNull && f.Value == nil {
			errMsg := "Filter Value is missing or nil"
			log.DB.Error("Failed to map user filters", errMsg, nil)
			return nil, errors.New(errMsg)
		}

		queryFilter[i] = UserFilter{
			Property: f.Property,
			Cond:     condMap[f.Cond],
			Value:    f.Value,
		}
	}

	log.DB.Trace("Mapping user filters: OK", nil)

	return queryFilter, nil
}
