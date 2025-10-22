package userfilterparser

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/core/filter"
	"sentinel/packages/core/user"
	FilterMapper "sentinel/packages/infrastructure/mappers/filter"
	parser "sentinel/packages/infrastructure/parsers"
	"strings"
	"time"
)

var prefixes = []string{
	string(user.IdProperty),
	string(user.LoginProperty),
	string(user.PasswordProperty),
	string(user.RolesProperty),
	string(user.DeletedAtProperty),
}

var cache = map[string]filter.Entity[user.Property]{}

// Each filter must be a string in a following format:
//
// <property>:<condition><value>
//
// Value should be omitted if condition is either "is null", either "is not null"
func Parse(rawFilter string) (filter.Entity[user.Property], *Error.Status) {
	parser.Log.Trace("Parsing user filter "+rawFilter+"...", nil)

	if cached, hit := cache[rawFilter]; hit {
		parser.Log.Trace("Cache hit: "+rawFilter, nil)
		return cached, nil
	}

	parser.Log.Trace("Cache miss: "+rawFilter, nil)

	var zero filter.Entity[user.Property]
	var property user.Property

	for _, pref := range prefixes {
		if strings.HasPrefix(rawFilter, pref) {
			property = user.Property(pref)
			break
		}
	}

	// if valid property wasn't found in prefix
	if property == "" {
		errMsg := "Filter does not begins with valid user property - " + rawFilter
		parser.Log.Error("Failed to parse user filter "+rawFilter, errMsg, nil)
		return zero, Error.NewStatusError(errMsg, http.StatusBadRequest)
	}

	// if condition doesn't begins with ':'
	if rawFilter[len(property)] != ':' {
		errMsg := "Syntax error: missing ':' before condition in filter - " + rawFilter
		parser.Log.Error("Failed to parse user filter "+rawFilter, errMsg, nil)
		return zero, Error.NewStatusError(errMsg, http.StatusBadRequest)
	}

	cond, err := FilterMapper.GetCondFromStringPrefix(rawFilter[len(property)+1:])
	if err != nil {
		return zero, Error.NewStatusError(err.Error(), http.StatusBadRequest)
	}

	if err := validatePropertyCond(property, cond); err != nil {
		parser.Log.Error("Failed to parse user filter "+rawFilter, err.Error(), nil)
		return zero, Error.NewStatusError(err.Error(), http.StatusBadRequest)
	}

	// cond is 100% valid, so there are will be no error in any case
	condStr, _ := FilterMapper.FormatCond(cond)

	var value any

	// 1 is ':'
	valueStart := len(property) + len(condStr) + 1

	switch property {
	case user.IdProperty, user.LoginProperty, user.PasswordProperty:
		value = rawFilter[valueStart:]
	case user.RolesProperty:
		value = strings.Split(strings.TrimSpace(rawFilter[valueStart:]), ",")
	case user.DeletedAtProperty:
		strTime := rawFilter[valueStart:]
		if len(strTime) == 0 { // if cond is either 'is null', either 'is not null'
			value = nil
			break
		}

		t, err := time.Parse(time.RFC3339, rawFilter[valueStart:])
		if err != nil {
			errMsg := "Filter has invalid time format (expected RFC3339) - " + rawFilter
			parser.Log.Error("Failed to parse user filter "+rawFilter, errMsg, nil)
			return zero, Error.NewStatusError(errMsg, http.StatusBadRequest)
		}
		value = t
	default:
		// property should be valid at this point, but an additional check won't be redundant
		// (especially when this function will need to be refactored/fixed/reworked)
		parser.Log.Panic(
			"Faield to parse user filter",
			"Unknown user property received: "+string(property),
			nil,
		)
		return zero, Error.StatusInternalError
	}

	f := filter.Entity[user.Property]{
		Property: property,
		Cond:     cond,
		Value:    value,
	}

	cache[rawFilter] = f

	parser.Log.Trace("Cache set: "+rawFilter, nil)

	parser.Log.Trace("Parsing user filter "+rawFilter+": OK", nil)

	return f, nil
}

var errorNoFilters = Error.NewStatusError(
	"At least one filter must be specified",
	http.StatusBadRequest,
)

// Calls Parse() function from this module for each element in rawFilter slice
func ParseAll(rawFilters []string) ([]filter.Entity[user.Property], *Error.Status) {
	filtersStr := strings.Join(rawFilters, ", ")
	parser.Log.Info("Parsing user filters: "+filtersStr+"...", nil)

	if rawFilters == nil || len(rawFilters) == 0 {
		return nil, errorNoFilters
	}

	filters := make([]filter.Entity[user.Property], len(rawFilters))

	for i, rawFilter := range rawFilters {
		filter, err := Parse(rawFilter)
		if err != nil {
			return nil, err
		}
		filters[i] = filter
	}

	parser.Log.Info("Parsing user filters: "+filtersStr+": OK", nil)

	return filters, nil
}
