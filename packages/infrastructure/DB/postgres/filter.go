package postgres

import (
	"fmt"
	"sentinel/packages/core/filter"
	"sentinel/packages/core/user"
	"strconv"
)

type filterCond string

const (
	condEqual 		   filterCond = "="
	condLess 		   filterCond = "<"
	condGreater 	   filterCond = ">"
	condLessOrEqual    filterCond = "<="
	condGreaterOrEqual filterCond = ">="
	condLike 		   filterCond = "LIKE"
	condIsNull		   filterCond = "IS NULL"
	condIsNotNull	   filterCond = "IS NOT NULL"
	condContains  	   filterCond = "@>"
	condContained  	   filterCond = "<@"
)

type QueryFilter struct {
	Property  user.Property
	Cond 	  filterCond
	Value 	  any
}

func (f QueryFilter) Build(n int) string {
	base := string(f.Property) + " " + string(f.Cond)

	// SELECT ... FROM ... WHERE <property> <cond> $n (omit if cond is "IS NULL" or "IS NOT NULL")
	if f.Cond == condIsNull || f.Cond == condIsNotNull {
		return base
	}

	return base + " $" + strconv.FormatInt(int64(n), 10)
}

func (f QueryFilter) StringValue() string {
	v, ok := f.Value.(string)
	if !ok {
		dbLogger.Panic(
			"Failed to find user",
			fmt.Sprintf("Query filter field 'value' has invalid type. Expected string, but got %T", f.Value),
			nil,
		)
		return ""
	}

	return v
}

var condMap = map[filter.Condition]filterCond {
	filter.Equal: condEqual,
	filter.Less: condLess,
	filter.Greater: condGreater,
	filter.LessOrEqual: condLessOrEqual,
	filter.GreaterOrEqual: condGreaterOrEqual,
	filter.Like: condLike,
	filter.IsNull: condIsNull,
	filter.IsNotNull: condIsNotNull,
	filter.Contains: condContains,
	filter.Containd: condContained,
}

func mapFilters(filters []filter.Entity[user.Property]) []QueryFilter {
	queryFilter := make([]QueryFilter, len(filters))

	for i, f := range filters {
		if f.Property == "" || f.Cond == 0 {
			dbLogger.Panic(
				"Failed to map entity filter to query filter",
				fmt.Sprintf(
					"Filter property or condition has invalid value. Property value: %v; Condition value: %v",
					f.Property, f.Cond,
				),
				nil,
			)
		}
		if f.Cond != filter.IsNull && f.Cond != filter.IsNotNull {
			if f.Value == nil {
				dbLogger.Panic(
					"Failed to map entity filter to query filter",
					fmt.Sprintf("Invalid filter value: %v", f.Value),
					nil,
				)
			}
		}

		queryFilter[i] = QueryFilter{
			Property: f.Property,
			Cond: condMap[f.Cond],
			Value: f.Value,
		}
	}

	return queryFilter
}

