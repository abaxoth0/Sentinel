package filtermapper

import (
	"errors"
	"fmt"
	"sentinel/packages/core/filter"
	"strings"
)

const (
	condEqualStr 		  = "="
	condLessStr     	  = "<"
	condGreaterStr		  = ">"
	condLessOrEqualStr	  = "<="
	condGreaterOrEqualStr = ">="
	condLikeStr			  = "~"
	condIsNullStr		  = "null"
	condIsNotNullStr 	  = "notnull"
	condContainsStr		  = "@>"
	condContainedStr	  = "<@"
)

var condsStrings = []string{
	condEqualStr,
	condLessStr,
	condGreaterStr,
	condLessOrEqualStr,
	condGreaterOrEqualStr,
	condLikeStr,
	condIsNullStr,
	condIsNotNullStr,
	condContainsStr,
	condContainedStr,
}

func GetCondFromStringPrefix(s string) (filter.Condition, error) {
	var cond filter.Condition

	for _, condStr := range condsStrings {
		if strings.HasPrefix(s, condStr) {
			cond = stringToCondMap[condStr]
			break
		}
	}

	if cond == 0 {
		return 0, errors.New("Failed to found valid filter condition: " + s)
	}

	return cond, nil
}

var condToStringMap = map[filter.Condition]string {
	filter.Equal: 		   condEqualStr,
	filter.Less: 		   condLessStr,
	filter.Greater: 	   condGreaterStr,
	filter.LessOrEqual:    condLessOrEqualStr,
	filter.GreaterOrEqual: condGreaterOrEqualStr,
	filter.Like: 		   condLikeStr,
	filter.IsNull: 		   condIsNullStr,
	filter.IsNotNull: 	   condIsNotNullStr,
	filter.Contains: 	   condContainsStr,
	filter.Containd: 	   condContainedStr,
}

var stringToCondMap = map[string]filter.Condition {
	condEqualStr:	   	   filter.Equal,
	condLessStr: 	   	   filter.Less,
	condGreaterStr: 	   filter.Greater,
	condLessOrEqualStr:    filter.LessOrEqual,
	condGreaterOrEqualStr: filter.GreaterOrEqual,
	condLikeStr: 	   	   filter.Like,
	condIsNullStr:    	   filter.IsNull,
	condIsNotNullStr: 	   filter.IsNotNull,
	condContainsStr: 	   filter.Contains,
	condContainedStr: 	   filter.Containd,
}

func ParseCond(rawCond string) (filter.Condition, error) {
	r := stringToCondMap[rawCond]
	if r == 0 {
		return 0, fmt.Errorf("Failed to parse filter condition '%s': no such condition", rawCond)
	}
	return r, nil
}

func FormatCond(cond filter.Condition) (string, error) {
	r := condToStringMap[cond]
	if r == "" {
		return "", fmt.Errorf("Failed to format filter condition with number '%d': no condition with such number", cond)
	}
	return r, nil
}

