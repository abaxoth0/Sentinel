package query

import "sentinel/packages/core/filter"

type filterCond string

const (
	CondEqual 		   filterCond = "="
	CondLess 		   filterCond = "<"
	CondGreater 	   filterCond = ">"
	CondLessOrEqual    filterCond = "<="
	CondGreaterOrEqual filterCond = ">="
	CondLike 		   filterCond = "LIKE"
	CondIsNull		   filterCond = "IS NULL"
	CondIsNotNull	   filterCond = "IS NOT NULL"
	CondContains  	   filterCond = "@>"
	CondContained  	   filterCond = "<@"
)

var condMap = map[filter.Condition]filterCond {
	filter.Equal: CondEqual,
	filter.Less: CondLess,
	filter.Greater: CondGreater,
	filter.LessOrEqual: CondLessOrEqual,
	filter.GreaterOrEqual: CondGreaterOrEqual,
	filter.Like: CondLike,
	filter.IsNull: CondIsNull,
	filter.IsNotNull: CondIsNotNull,
	filter.Contains: CondContains,
	filter.Containd: CondContained,
}

