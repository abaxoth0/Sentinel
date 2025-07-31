package filtermapper

import (
	"errors"
	"fmt"
	"sentinel/packages/core/filter"
	mapper "sentinel/packages/infrastructure/mappers"
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
	mapper.Log.Trace("Getting condition from sting prefix: "+s+"...", nil)

	var cond filter.Condition

	for _, condStr := range condsStrings {
		if strings.HasPrefix(s, condStr) {
			cond = stringToCondMap[condStr]
			break
		}
	}

	if cond == 0 {
		errMsg := "Failed to found valid filter condition: " + s
		mapper.Log.Error("Failed to get condition from sting prefix: "+s, errMsg, nil)
		return 0, errors.New(errMsg)
	}

	mapper.Log.Trace("Getting condition from sting prefix: "+s+": OK", nil)

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
	mapper.Log.Trace("Parsing filter condition: "+rawCond+"...", nil)
	r := stringToCondMap[rawCond]
	if r == 0 {
		errMsg := "No such condition"
		mapper.Log.Error("Parsing filter condition: "+rawCond, errMsg, nil)
		return 0, errors.New(errMsg)
	}
	mapper.Log.Trace("Parsing filter condition: "+rawCond+": OK", nil)
	return r, nil
}

func FormatCond(cond filter.Condition) (string, error) {
	mapper.Log.Trace("Formatting filter condition into the string...", nil)
	r := condToStringMap[cond]
	if r == "" {
		errMsg := fmt.Sprintf("There are no condition with number '%d'", cond)
		mapper.Log.Error("Failed to format filter condition into the string", errMsg, nil)
		return "", errors.New(errMsg)
	}
	mapper.Log.Trace("Formatting filter condition into the string: OK", nil)
	return r, nil
}

