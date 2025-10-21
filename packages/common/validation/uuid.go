package validation

import (
	Error "sentinel/packages/common/errors"
	"strings"

	"github.com/google/uuid"
)

// Returns nil if 'v' is valid uuid,
// otherwise returns either Error.NoValue or Error.InvalidValue.
func UUID(v string) *Error.Validation {
    if strings.ReplaceAll(v, " ", "") == "" {
        return Error.NoValue
    }

    if err := uuid.Validate(v); err != nil {
        return Error.InvalidValue
    }

    return nil
}
